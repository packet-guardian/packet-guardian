// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package common

import (
	"context"
	"net"
	"net/http"
	"strings"
)

// Environment

// GetEnvironmentFromContext retrieves the Environment from the current request.
func GetEnvironmentFromContext(r *http.Request) *Environment {
	if rv := r.Context().Value(SessionEnvKey); rv != nil {
		return rv.(*Environment)
	}
	return nil
}

// SetEnvironmentToContext sets an Environment for the current request.
func SetEnvironmentToContext(r *http.Request, e *Environment) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), SessionEnvKey, e))
}

// Session

// GetSessionFromContext retrieves the Session from the current request.
func GetSessionFromContext(r *http.Request) *Session {
	if rv := r.Context().Value(SessionKey); rv != nil {
		return rv.(*Session)
	}
	return nil
}

// SetSessionToContext sets an Session for the current request.
func SetSessionToContext(r *http.Request, s *Session) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), SessionKey, s))
}

// IP of client

// GetIPFromContext retrieves the IP address from the current request.
func GetIPFromContext(r *http.Request) net.IP {
	if rv := r.Context().Value(SessionIPKey); rv != nil {
		return rv.(net.IP)
	}
	return nil
}

// SetIPToContext sets an IP address for the current request.
func SetIPToContext(r *http.Request) *http.Request {
	return r.WithContext(context.WithValue(
		r.Context(),
		SessionIPKey,
		net.ParseIP(strings.Split(r.RemoteAddr, ":")[0]),
	))
}
