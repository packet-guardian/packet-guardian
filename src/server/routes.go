// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"runtime"
	"strings"

	"github.com/gorilla/mux"
	"github.com/usi-lfkeitel/packet-guardian/src/auth"
	"github.com/usi-lfkeitel/packet-guardian/src/common"
	"github.com/usi-lfkeitel/packet-guardian/src/controllers"
	"github.com/usi-lfkeitel/packet-guardian/src/controllers/api"
	"github.com/usi-lfkeitel/packet-guardian/src/models"
	mid "github.com/usi-lfkeitel/packet-guardian/src/server/middleware"
	"github.com/usi-lfkeitel/pg-dhcp"
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

	guestController := controllers.NewGuestController(e)
	r.Handle("/register/guest", mid.CheckReg(e, http.HandlerFunc(guestController.RegistrationHandler))).Methods("GET", "POST")
	r.Handle("/register/guest/verify", mid.CheckReg(e, http.HandlerFunc(guestController.VerificationHandler))).Methods("GET", "POST")

	r.PathPrefix("/admin").Handler(adminRouter(e))
	r.PathPrefix("/api").Handler(apiRouter(e))

	// Development Routes
	if e.IsDev() {
		r.PathPrefix("/dev").Handler(devRouter(e))
		r.PathPrefix("/debug").Handler(debugRouter(e))
		e.Log.Debug("Profiling enabled")
	}

	h := mid.BlacklistCheck(e, r) // Enforce a blacklist check
	h = mid.Cache(e, h)           // Set cache headers if needed
	h = mid.SetSessionInfo(e, h)  // Adds Environment and user information to requet context
	h = mid.Logging(e, h)         // Logging
	h = mid.Panic(e, h)           // Panic catcher

	return h
}

func devRouter(e *common.Environment) http.Handler {
	r := mux.NewRouter()
	r.NotFoundHandler = http.HandlerFunc(notFoundHandler)

	devController := controllers.NewDevController(e)
	r.HandleFunc("/dev/reloadtemp", devController.ReloadTemplates).Methods("GET")
	r.HandleFunc("/dev/reloadconf", devController.ReloadConfiguration).Methods("GET")

	h := mid.CheckAdmin(r)
	h = mid.CheckAuth(h)
	return h
}

func debugRouter(e *common.Environment) http.Handler {
	r := mux.NewRouter()
	r.NotFoundHandler = http.HandlerFunc(notFoundHandler)

	r.HandleFunc("/debug/pprof", pprof.Index)
	r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("/debug/pprof/profile", pprof.Profile)
	r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	r.HandleFunc("/debug/pprof/trace", pprof.Trace)
	// Manually add support for paths linked to by index page at /debug/pprof/
	r.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	r.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	r.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	r.Handle("/debug/pprof/block", pprof.Handler("block"))

	r.HandleFunc("/debug/heap-stats", heapStats)

	h := mid.CheckAdmin(r)
	h = mid.CheckAuth(h)
	return h
}

func heapStats(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Fprintf(w,
		"HeapSys: %d, HeapAlloc: %d, HeapIdle: %d, HeapReleased: %d\n",
		m.HeapSys,
		m.HeapAlloc,
		m.HeapIdle,
		m.HeapReleased,
	)
}

func adminRouter(e *common.Environment) http.Handler {
	r := mux.NewRouter()
	r.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	get := r.Methods("GET").Subrouter()

	adminController := controllers.NewAdminController(e)
	get.HandleFunc("/admin", adminController.DashboardHandler)
	get.HandleFunc("/admin/search", adminController.SearchHandler)
	get.HandleFunc("/admin/manage/user/{username}", adminController.ManageHandler)
	get.HandleFunc("/admin/manage/device/{mac:(?:[0-f]{2}:?){6}}", adminController.ShowDeviceHandler)
	get.HandleFunc("/admin/users", adminController.AdminUserListHandler)
	get.HandleFunc("/admin/users/{username}", adminController.AdminUserHandler)
	get.HandleFunc("/admin/reports", adminController.ReportHandler)
	get.HandleFunc("/admin/reports/{report}", adminController.ReportHandler)

	h := mid.CheckAdmin(r)
	h = mid.CheckAuth(h)
	return h
}

func apiRouter(e *common.Environment) http.Handler {
	r := mux.NewRouter()

	deviceApiController := api.NewDeviceController(e)
	r.HandleFunc("/api/device", deviceApiController.RegistrationHandler).Methods("POST")
	r.HandleFunc("/api/device/{username}", deviceApiController.DeleteHandler).Methods("DELETE")
	r.HandleFunc("/api/device/_reassign", deviceApiController.ReassignHandler).Methods("POST")

	blacklistController := api.NewBlacklistController(e)
	r.HandleFunc("/api/blacklist/user/{username}", blacklistController.BlacklistUserHandler).Methods("POST", "DELETE")
	r.HandleFunc("/api/blacklist/device/{username}", blacklistController.BlacklistDeviceHandler).Methods("POST", "DELETE")

	userApiController := api.NewUserController(e)
	r.HandleFunc("/api/user", userApiController.UserHandler).Methods("POST", "DELETE")

	return mid.CheckAuth(r)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if auth.IsLoggedIn(r) {
		sessionUser := models.GetUserFromContext(r)
		if sessionUser.Can(models.ViewAdminPage) {
			http.Redirect(w, r, "/admin", http.StatusTemporaryRedirect)
			return
		}

		http.Redirect(w, r, "/manage", http.StatusTemporaryRedirect)
		return
	}

	e := common.GetEnvironmentFromContext(r)
	ip := net.ParseIP(strings.Split(r.RemoteAddr, ":")[0])
	reg, err := dhcp.IsRegisteredByIP(models.NewLeaseStore(e), ip)
	if err != nil {
		e.Log.WithField("Err", err).Notice("Couldn't get registration status")
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	if reg {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, "/register", http.StatusTemporaryRedirect)
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api") {
		common.NewEmptyAPIResponse().WriteResponse(w, http.StatusNotFound)
		return
	}

	sessionUser := models.GetUserFromContext(r)
	if sessionUser.Can(models.ViewAdminPage) {
		http.Redirect(w, r, "/admin", http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
}
