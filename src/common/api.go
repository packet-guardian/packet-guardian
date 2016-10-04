// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package common

import (
	"encoding/json"
	"net/http"

	"github.com/lfkeitel/verbose"
)

// A APIResponse is returned as a JSON struct to the client
type APIResponse struct {
	Message string
	Data    interface{}
}

// NewAPIResponse creates an APIResponse object with status c, message m, and data d
func NewAPIResponse(m string, d interface{}) *APIResponse {
	return &APIResponse{
		Message: m,
		Data:    d,
	}
}

func NewEmptyAPIResponse() *APIResponse {
	return &APIResponse{}
}

// Encode the APIResponse into JSON
func (a *APIResponse) Encode() []byte {
	b, err := json.Marshal(a)
	if err != nil {
		SystemLogger.WithFields(verbose.Fields{
			"error":   err,
			"package": "common",
		}).Error("Error encoding API response data")
	}
	return b
}

func (a *APIResponse) WriteResponse(w http.ResponseWriter, code int) (int64, error) {
	r := a.Encode()
	l := len(r)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	if code == http.StatusNoContent {
		return 0, nil
	}
	w.Write(r)
	return int64(l), nil
}
