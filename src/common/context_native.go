// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// +build go1.7

package common

import (
	"context"
	"net/http"
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
