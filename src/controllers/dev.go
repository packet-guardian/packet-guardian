// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package controllers

import (
	"net/http"

	"github.com/usi-lfkeitel/packet-guardian/src/common"
)

type Dev struct {
	e *common.Environment
}

func NewDevController(e *common.Environment) *Dev {
	return &Dev{e: e}
}

// Dev mode route handlers
func (d *Dev) ReloadTemplates(w http.ResponseWriter, r *http.Request) {
	if err := d.e.Views.Reload(); err != nil {
		w.Write([]byte("Error loading HTML templates: " + err.Error()))
		return
	}
	w.Write([]byte("Templates reloaded"))
}
