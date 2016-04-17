package api

import (
	"net/http"

	"github.com/onesimus-systems/packet-guardian/src/common"
)

type User struct {
	e *common.Environment
}

func NewUserController(e *common.Environment) *User {
	return &User{e: e}
}

func (u *User) adminUserHandler(w http.ResponseWriter, r *http.Request) {

}

func (u *User) saveUserHandler(w http.ResponseWriter, r *http.Request) {

	common.NewAPIOK("User created", nil).WriteTo(w)
}

func (u *User) deleteUserHandler(w http.ResponseWriter, r *http.Request) {

	common.NewAPIOK("User deleted", nil).WriteTo(w)
}
