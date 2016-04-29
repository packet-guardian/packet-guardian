package server

import (
	"net"
	"net/http"
	"net/http/pprof"
	"strings"

	"github.com/gorilla/mux"
	"github.com/onesimus-systems/packet-guardian/src/auth"
	"github.com/onesimus-systems/packet-guardian/src/common"
	"github.com/onesimus-systems/packet-guardian/src/controllers"
	"github.com/onesimus-systems/packet-guardian/src/controllers/api"
	"github.com/onesimus-systems/packet-guardian/src/dhcp"
	"github.com/onesimus-systems/packet-guardian/src/models"
	mid "github.com/onesimus-systems/packet-guardian/src/server/middleware"
)

func LoadRoutes(e *common.Environment) http.Handler {
	r := mux.NewRouter().StrictSlash(true)

	// Page routes
	r.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	r.HandleFunc("/", rootHandler)
	r.PathPrefix("/public").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("./public/"))))

	authController := controllers.NewAuthController(e)
	r.HandleFunc("/login", authController.LoginHandler).Methods("GET", "POST")
	r.HandleFunc("/logout", authController.LogoutHandler).Methods("GET")

	manageController := controllers.NewManagerController(e)
	r.HandleFunc("/register", manageController.RegistrationHandler).Methods("GET")
	r.Handle("/manage", mid.CheckAuth(http.HandlerFunc(manageController.ManageHandler))).Methods("GET")

	r.PathPrefix("/admin").Handler(adminRouter(e))
	r.PathPrefix("/api").Handler(apiRouter(e))

	// Development Routes
	if e.Dev {
		devController := controllers.NewDevController(e)
		s := r.PathPrefix("/dev").Subrouter()
		s.HandleFunc("/reloadtemp", devController.ReloadTemplates).Methods("GET")
		s.HandleFunc("/reloadconf", devController.ReloadConfiguration).Methods("GET")

		// Add routes for profiler
		s = r.PathPrefix("/debug").Subrouter()
		s.HandleFunc("/pprof/", pprof.Index)
		s.HandleFunc("/pprof/cmdline", pprof.Cmdline)
		s.HandleFunc("/pprof/profile", pprof.Profile)
		s.HandleFunc("/pprof/symbol", pprof.Symbol)
		s.HandleFunc("/pprof/trace", pprof.Trace)
		// Manually add support for paths linked to by index page at /debug/pprof/
		s.Handle("/pprof/goroutine", pprof.Handler("goroutine"))
		s.Handle("/pprof/heap", pprof.Handler("heap"))
		s.Handle("/pprof/threadcreate", pprof.Handler("threadcreate"))
		s.Handle("/pprof/block", pprof.Handler("block"))
		e.Log.Debug("Profiling enabled")
	}

	h := mid.Cache(e, r)         // Set cache headers if needed
	h = mid.SetSessionInfo(e, h) // Adds Environment and user information to requet context
	h = mid.Logging(e, h)        // Logging

	return h
}

func adminRouter(e *common.Environment) http.Handler {
	r := mux.NewRouter()
	r.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	get := r.Methods("GET").Subrouter()

	adminController := controllers.NewAdminController(e)
	get.HandleFunc("/admin", adminController.DashboardHandler)
	get.HandleFunc("/admin/search", adminController.SearchHandler)
	get.HandleFunc("/admin/manage/{username}", adminController.ManageHandler)
	get.HandleFunc("/admin/users", adminController.AdminUserListHandler)
	get.HandleFunc("/admin/users/{username}", adminController.AdminUserHandler)

	h := mid.CheckAdmin(r)
	h = mid.CheckAuth(h)
	return h
}

func apiRouter(e *common.Environment) http.Handler {
	r := mux.NewRouter()

	deviceApiController := api.NewDeviceController(e)
	r.HandleFunc("/api/device/register", deviceApiController.RegistrationHandler).Methods("POST")
	r.HandleFunc("/api/device/delete", deviceApiController.DeleteHandler).Methods("DELETE")

	blacklistController := api.NewBlacklistController(e)
	r.HandleFunc("/api/blacklist/{type}", blacklistController.BlacklistHandler).Methods("POST", "DELETE")

	userApiController := api.NewUserController(e)
	r.HandleFunc("/api/user", userApiController.UserHandler).Methods("POST", "DELETE")

	return mid.CheckAPI(r)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	e := common.GetEnvironmentFromContext(r)
	ip := strings.Split(r.RemoteAddr, ":")[0]
	reg, err := dhcp.IsRegisteredByIP(e, net.ParseIP(ip))
	if err != nil {
		e.Log.Errorf("Error checking auto registration IP: %s", err.Error())
	}

	if auth.IsLoggedIn(r) {
		sessionUser := models.GetUserFromContext(r)
		if sessionUser.IsHelpDesk() || sessionUser.IsAdmin() {
			http.Redirect(w, r, "/admin", http.StatusTemporaryRedirect)
		} else {
			http.Redirect(w, r, "/manage", http.StatusTemporaryRedirect)
		}
		return
	}

	if reg {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
	} else {
		http.Redirect(w, r, "/register", http.StatusTemporaryRedirect)
	}
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	e := common.GetEnvironmentFromContext(r)
	e.Log.GetLogger("server").Infof("Path not found %s", r.RequestURI)
	sessionUser := models.GetUserFromContext(r)
	if sessionUser.IsHelpDesk() || sessionUser.IsAdmin() {
		http.Redirect(w, r, "/admin", http.StatusTemporaryRedirect)
	} else {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
	}
}
