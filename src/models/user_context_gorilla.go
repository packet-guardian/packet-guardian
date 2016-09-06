// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// +build !go1.7

package models

import (
	"net/http"

	"github.com/gorilla/context"
	"github.com/usi-lfkeitel/packet-guardian/src/common"
)

func GetUserFromContext(r *http.Request) *User {
	if rv := context.Get(r, common.SessionUserKey); rv != nil {
		return rv.(*User)
	}
	return nil
}

func SetUserToContext(r *http.Request, u *User) *http.Request {
	context.Set(r, common.SessionUserKey, u)
	return r
}
