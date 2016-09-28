// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// +build go1.7

package common

import (
	"context"
	"net"
	"net/http"
	"strings"
)

// Environment

func GetEnvironmentFromContext(r *http.Request) *Environment {
	if rv := r.Context().Value(SessionEnvKey); rv != nil {
		return rv.(*Environment)
	}
	return nil
}

func SetEnvironmentToContext(r *http.Request, e *Environment) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), SessionEnvKey, e))
}

// Session

func GetSessionFromContext(r *http.Request) *Session {
	if rv := r.Context().Value(SessionKey); rv != nil {
		return rv.(*Session)
	}
	return nil
}

func SetSessionToContext(r *http.Request, s *Session) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), SessionKey, s))
}

// IP of client

func GetIPFromContext(r *http.Request) net.IP {
	if rv := r.Context().Value(SessionIPKey); rv != nil {
		return rv.(net.IP)
	}
	return nil
}

func SetIPToContext(r *http.Request) *http.Request {
	return r.WithContext(context.WithValue(
		r.Context(),
		SessionIPKey,
		net.ParseIP(strings.Split(r.RemoteAddr, ":")[0]),
	))
}
