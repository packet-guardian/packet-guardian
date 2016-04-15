package controllers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/onesimus-systems/packet-guardian/src/auth"
	"github.com/onesimus-systems/packet-guardian/src/common"
)

type Auth struct {
	e *common.Environment
}

func NewAuthController(e *common.Environment) *Auth {
	return &Auth{e: e}
}

func (a *Auth) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/login", a.loginHandler).Methods("GET", "POST")
	r.HandleFunc("/logout", a.logoutHandler).Methods("GET")
}

func (a *Auth) loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		loggedin := a.e.Sessions.GetSession(r).GetBool("loggedin", false)
		if loggedin {
			http.Redirect(w, r, "/manage", http.StatusTemporaryRedirect)
		}
		data := struct{ FlashMessage string }{}
		if err := a.e.Views.NewView("login").Render(w, data); err != nil {
			a.e.Log.Error(err.Error())
		}
	} else if r.Method == "POST" {
		// Assume invalid until convinced otherwise
		resp := common.NewAPIResponse(common.APIStatusInvalidAuth, "Invalid login", nil)
		username := r.FormValue("username")
		if auth.IsValidLogin(a.e.DB, username, r.FormValue("password")) {
			resp.Code = common.APIStatusOK
			resp.Message = ""
			sess := a.e.Sessions.GetSession(r)
			sess.Set("loggedin", true)
			sess.Set("username", username)
			sess.Save(r, w)
			if common.StringInSlice(username, a.e.Config.Auth.AdminUsers) {
				a.e.Log.Infof("Successful login by administrative user %s", username)
			} else {
				a.e.Log.Infof("Successful login by user %s", username)
			}
		} else {
			a.e.Log.Errorf("Incorrect login from user %s", username)
		}
		resp.WriteTo(w)
	}
}

// LogoutHandler voids a user's session
func (a *Auth) logoutHandler(w http.ResponseWriter, r *http.Request) {
	sess := a.e.Sessions.GetSession(r)
	auth.LogoutUser(a.e, w, r)
	a.e.Log.Infof("Successful logout by user %s", sess.GetString("username", "[unknown]"))
	http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
}
