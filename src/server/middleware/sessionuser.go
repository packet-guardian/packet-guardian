package middleware

import (
	"net/http"

	"github.com/gorilla/context"
	"github.com/onesimus-systems/packet-guardian/src/common"
	"github.com/onesimus-systems/packet-guardian/src/models"
)

func SetSessionUser(e *common.Environment, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := e.Sessions.GetSession(r).GetString("username")
		sessionUser, err := models.GetUserByUsername(e, username)
		if err != nil {
			e.Log.Error("Failed to get session user: " + err.Error())
		}
		context.Set(r, "sessionUser", sessionUser)
		next(w, r)
	}
}
