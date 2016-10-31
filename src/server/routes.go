// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"fmt"
	"net/http"
	"net/http/pprof"
	"runtime"
	"strings"

	"github.com/dchest/captcha"
	"github.com/gorilla/context"
	"github.com/julienschmidt/httprouter"
	"github.com/lfkeitel/verbose"

	"github.com/usi-lfkeitel/packet-guardian/src/auth"
	"github.com/usi-lfkeitel/packet-guardian/src/common"
	"github.com/usi-lfkeitel/packet-guardian/src/controllers"
	"github.com/usi-lfkeitel/packet-guardian/src/controllers/api"
	"github.com/usi-lfkeitel/packet-guardian/src/models"
	mid "github.com/usi-lfkeitel/packet-guardian/src/server/middleware"
	"github.com/usi-lfkeitel/pg-dhcp"
)

func LoadRoutes(e *common.Environment) http.Handler {
	r := httprouter.New()
	r.NotFound = http.HandlerFunc(notFoundHandler)

	r.Handler("GET", "/", midStack(e, http.HandlerFunc(rootHandler)))
	r.ServeFiles("/public/*filepath", http.Dir("./public"))

	authController := controllers.NewAuthController(e)
	r.Handler("GET", "/login", midStack(e, http.HandlerFunc(authController.LoginHandler)))
	r.Handler("POST", "/login", midStack(e, http.HandlerFunc(authController.LoginHandler)))
	r.Handler("GET", "/logout", midStack(e, http.HandlerFunc(authController.LogoutHandler)))

	manageController := controllers.NewManagerController(e)
	r.Handler("GET", "/register", midStack(e, http.HandlerFunc(manageController.RegistrationHandler)))
	r.Handler("GET", "/manage", midStack(e, mid.CheckAuth(http.HandlerFunc(manageController.ManageHandler))))

	guestController := controllers.NewGuestController(e)
	r.Handler("GET", "/register/guest", midStack(e, mid.CheckGuestReg(e,
		http.HandlerFunc(guestController.RegistrationHandler))))
	r.Handler("POST", "/register/guest", midStack(e, mid.CheckGuestReg(e,
		http.HandlerFunc(guestController.RegistrationHandler))))
	r.Handler("GET", "/register/guest/verify", midStack(e, mid.CheckGuestReg(e,
		http.HandlerFunc(guestController.VerificationHandler))))
	r.Handler("POST", "/register/guest/verify", midStack(e, mid.CheckGuestReg(e,
		http.HandlerFunc(guestController.VerificationHandler))))

	r.Handler("GET", "/admin/*a", midStack(e, adminRouter(e)))
	r.Handler("POST", "/api/*a", midStack(e, apiRouter(e)))
	r.Handler("DELETE", "/api/*a", midStack(e, apiRouter(e)))

	r.Handler("GET", "/captcha/*a", captcha.Server(captcha.StdWidth, captcha.StdHeight))

	if e.IsDev() {
		r.Handler("GET", "/dev/*a", midStack(e, devRouter(e)))
		r.Handler("GET", "/debug/*a", midStack(e, debugRouter(e)))
		e.Log.Debug("Profiling enabled")
	}

	h := mid.Logging(e, r) // Logging
	h = mid.Panic(e, h)    // Panic catcher
	return h
}

func midStack(e *common.Environment, h http.Handler) http.Handler {
	h = mid.BlacklistCheck(e, h) // Enforce a blacklist check
	h = mid.Cache(e, h)          // Set cache headers if needed
	h = mid.SetSessionInfo(e, h) // Adds Environment and user information to requet context
	h = context.ClearHandler(h)  // Clear Gorilla sessions
	return h
}

func devRouter(e *common.Environment) http.Handler {
	r := httprouter.New()
	r.NotFound = http.HandlerFunc(notFoundHandler)

	devController := controllers.NewDevController(e)
	r.HandlerFunc("GET", "/dev/reloadtemp", devController.ReloadTemplates)
	r.HandlerFunc("GET", "/dev/reloadconf", devController.ReloadConfiguration)

	h := mid.CheckAdmin(r)
	h = mid.CheckAuth(h)
	return h
}

func debugRouter(e *common.Environment) http.Handler {
	r := httprouter.New()
	r.NotFound = http.HandlerFunc(notFoundHandler)

	r.HandlerFunc("GET", "/debug/pprof", pprof.Index)
	r.HandlerFunc("GET", "/debug/pprof/cmdline", pprof.Cmdline)
	r.HandlerFunc("GET", "/debug/pprof/profile", pprof.Profile)
	r.HandlerFunc("GET", "/debug/pprof/symbol", pprof.Symbol)
	r.HandlerFunc("GET", "/debug/pprof/trace", pprof.Trace)
	// Manually add support for paths linked to by index page at /debug/pprof/
	r.Handler("GET", "/debug/pprof/goroutine", pprof.Handler("goroutine"))
	r.Handler("GET", "/debug/pprof/heap", pprof.Handler("heap"))
	r.Handler("GET", "/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	r.Handler("GET", "/debug/pprof/block", pprof.Handler("block"))

	r.HandlerFunc("GET", "/debug/heap-stats", heapStats)

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
	r := httprouter.New()
	r.NotFound = http.HandlerFunc(notFoundHandler)

	adminController := controllers.NewAdminController(e)
	r.GET("/admin/", adminController.DashboardHandler)
	r.GET("/admin/search", adminController.SearchHandler)
	r.GET("/admin/manage/user/:username", adminController.ManageHandler)
	r.GET("/admin/manage/device/:mac", adminController.ShowDeviceHandler)
	r.GET("/admin/users", adminController.AdminUserListHandler)
	r.GET("/admin/users/:username", adminController.AdminUserHandler)
	r.GET("/admin/reports", adminController.ReportHandler)
	r.GET("/admin/reports/:report", adminController.ReportHandler)

	h := mid.CheckAdmin(r)
	h = mid.CheckAuth(h)
	return h
}

func apiRouter(e *common.Environment) http.Handler {
	r := httprouter.New()

	deviceApiController := api.NewDeviceController(e)
	r.POST("/api/device", deviceApiController.RegistrationHandler)
	r.DELETE("/api/device/user/:username", deviceApiController.DeleteHandler)
	r.POST("/api/device/reassign", deviceApiController.ReassignHandler)
	r.POST("/api/device/mac/:mac/description", deviceApiController.EditDescriptionHandler)
	r.POST("/api/device/mac/:mac/expiration", deviceApiController.EditExpirationHandler)

	blacklistController := api.NewBlacklistController(e)
	r.POST("/api/blacklist/user/:username", blacklistController.BlacklistUserHandler)
	r.DELETE("/api/blacklist/user/:username", blacklistController.BlacklistUserHandler)

	r.POST("/api/blacklist/device/:username", blacklistController.BlacklistDeviceHandler)
	r.DELETE("/api/blacklist/device/:username", blacklistController.BlacklistDeviceHandler)

	userApiController := api.NewUserController(e)
	r.POST("/api/user", userApiController.UserHandler)
	r.DELETE("/api/user", userApiController.UserHandler)

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
	reg, err := dhcp.IsRegisteredByIP(models.GetLeaseStore(e), common.GetIPFromContext(r))
	if err != nil {
		e.Log.WithFields(verbose.Fields{
			"error":   err,
			"package": "routes",
			"ip":      common.GetIPFromContext(r).String(),
		}).Error("Error getting registration status")
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
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
