// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// +build go1.7

package models

import (
	"context"
	"net/http"

	"github.com/usi-lfkeitel/packet-guardian/src/common"
)

func GetUserFromContext(r *http.Request) *User {
	if rv := r.Context().Value(common.SessionUserKey); rv != nil {
		return rv.(*User)
	}
	return nil
}

func SetUserToContext(r *http.Request, u *User) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), common.SessionUserKey, u))
}
