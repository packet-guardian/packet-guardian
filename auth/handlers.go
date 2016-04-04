package auth

import (
	"net/http"

	"github.com/onesimus-systems/packet-guardian/common"
)

// LoginHandler handles a login POST request
func LoginHandler(e *common.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Assume invalid until convinced otherwise
		resp := common.NewAPIResponse(common.APIStatusInvalidAuth, "Invalid login", nil)
		if IsValidLogin(e.DB, r.FormValue("username"), r.FormValue("password")) {
			resp.Code = common.APIStatusOK
			resp.Message = ""
			sess := e.Sessions.GetSession(r, e.Config.Webserver.SessionName)
			sess.Set("loggedin", true)
			sess.Set("username", r.FormValue("username"))
			sess.Save(r, w)
		}
		resp.WriteTo(w)
	}
}

func LoginPageHandler(e *common.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess := e.Sessions.GetSession(r, e.Config.Webserver.SessionName)
		loggedin := sess.GetBool("loggedin", false)
		if loggedin {
			http.Redirect(w, r, "/manage", http.StatusTemporaryRedirect)
		}
		data := struct {
			SiteTitle   string
			CompanyName string
		}{
			SiteTitle:   e.Config.Core.SiteTitle,
			CompanyName: e.Config.Core.SiteCompanyName,
		}
		e.Templates.ExecuteTemplate(w, "login", data)
	}
}

func Logout(e *common.Environment, w http.ResponseWriter, r *http.Request) {
	sess := e.Sessions.GetSession(r, e.Config.Webserver.SessionName)
	sess.Set("loggedin", false)
	sess.Save(r, w)
}

func IsLoggedIn(e *common.Environment, r *http.Request) bool {
	sess := e.Sessions.GetSession(r, e.Config.Webserver.SessionName)
	return sess.GetBool("loggedin", false)
}

func CheckAuth(e *common.Environment, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !IsLoggedIn(e, r) {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}
		next(w, r)
	}
}

func CheckAuthAPI(e *common.Environment, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !IsLoggedIn(e, r) {
			common.NewAPIResponse(common.APIStatusAuthNeeded, "Not logged in", nil).WriteTo(w)
			return
		}
		next(w, r)
	}
}
