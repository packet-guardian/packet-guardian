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
			if common.StringInSlice(username, e.Config.Auth.AdminUsers) {
				e.Log.Infof("Successful login by administrative user %s", username)
			} else {
				e.Log.Infof("Successful login by user %s", username)
			}
		} else {
			e.Log.Errorf("Incorrect login from user %s", username)
		}
		resp.WriteTo(w)
	}
}

// LogoutHandler voids a user's session
func LogoutHandler(e *common.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess := e.Sessions.GetSession(r, e.Config.Webserver.SessionName)
		LogoutUser(e, w, r)
		e.Log.Infof("Successful logout by user %s", sess.GetString("username", "[unknown]"))
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
	}
}

// LoginPageHandler will either redirect to the manage page if logged in or will show a login box
func LoginPageHandler(e *common.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		loggedin := e.Sessions.GetSession(r, e.Config.Webserver.SessionName).GetBool("loggedin", false)
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

// CheckAuthMid is middleware to check if a user is logged in, if not it will redirect to the login page
func CheckAuthMid(e *common.Environment, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !IsLoggedIn(e, r) {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}
		next(w, r)
	}
}

// CheckAuthAPIMid is middleware to check if a user is logged in, if not it will return an AuthNeeded api status
func CheckAuthAPIMid(e *common.Environment, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !IsLoggedIn(e, r) {
			common.NewAPIResponse(common.APIStatusAuthNeeded, "Not logged in", nil).WriteTo(w)
			return
		}
		next(w, r)
	}
}

// CheckAdminMid is middleware that checks if a user is an administrator, it calls
// the CheckAuthMid middleware before checking itself
func CheckAdminMid(e *common.Environment, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		adminMid := func(w http.ResponseWriter, r *http.Request) {
			if !IsAdminUser(e, r) {
				http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
				return
			}
			next(w, r)
		}
		CheckAuthMid(e, adminMid)(w, r)
	}
}
