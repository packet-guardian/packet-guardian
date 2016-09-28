// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// +build !go1.7

package common

import (
	"net"
	"net/http"
	"strings"

	"github.com/gorilla/context"
)

// Environment

func GetEnvironmentFromContext(r *http.Request) *Environment {
	if rv := context.Get(r, SessionEnvKey); rv != nil {
		return rv.(*Environment)
	}
	return nil
}

func SetEnvironmentToContext(r *http.Request, e *Environment) *http.Request {
	context.Set(r, SessionEnvKey, e)
	return r
}

// Session

func GetSessionFromContext(r *http.Request) *Session {
	if rv := context.Get(r, SessionKey); rv != nil {
		return rv.(*Session)
	}
	return nil
}

func SetSessionToContext(r *http.Request, s *Session) *http.Request {
	context.Set(r, SessionKey, s)
	return r
}

// IP of client

func GetIPFromContext(r *http.Request) net.IP {
	if rv := context.Get(r, SessionIPKey); rv != nil {
		return rv.(net.IP)
	}
	return nil
}

func SetIPToContext(r *http.Request) *http.Request {
	context.Set(r, SessionIPKey, net.ParseIP(strings.Split(r.RemoteAddr, ":")[0]))
	return r
}
