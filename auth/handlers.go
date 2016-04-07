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
		username := r.FormValue("username")
		if IsValidLogin(e.DB, username, r.FormValue("password")) {
			resp.Code = common.APIStatusOK
			resp.Message = ""
			sess := e.Sessions.GetSession(r, e.Config.Webserver.SessionName)
			sess.Set("loggedin", true)
			sess.Set("username", username)
			sess.Save(r, w)
			e.Log.Infof("Successful login by user %s", username)
		} else {
			e.Log.Errorf("Incorrect login from user %s", username)
		}
		resp.WriteTo(w)
	}
}

func LogoutHandler(e *common.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess := e.Sessions.GetSession(r, e.Config.Webserver.SessionName)
		LogoutUser(e, w, r)
		e.Log.Infof("Successful logout by user %s", sess.GetString("username", "[unknown]"))
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
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

func LogoutUser(e *common.Environment, w http.ResponseWriter, r *http.Request) {
	sess := e.Sessions.GetSession(r, e.Config.Webserver.SessionName)
	sess.Set("loggedin", false)
	sess.Delete(r, w)
}

func IsLoggedIn(e *common.Environment, r *http.Request) bool {
	sess := e.Sessions.GetSession(r, e.Config.Webserver.SessionName)
	return sess.GetBool("loggedin", false)
}

func CheckAuthMid(e *common.Environment, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !IsLoggedIn(e, r) {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}
		next(w, r)
	}
}

func CheckAuthAPIMid(e *common.Environment, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !IsLoggedIn(e, r) {
			common.NewAPIResponse(common.APIStatusAuthNeeded, "Not logged in", nil).WriteTo(w)
			return
		}
		next(w, r)
	}
}
