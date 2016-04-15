package controllers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/onesimus-systems/packet-guardian/src/common"
)

type Dev struct {
	e *common.Environment
}

func NewDevController(e *common.Environment) *Dev {
	return &Dev{e: e}
}

func (d *Dev) RegisterRoutes(r *mux.Router) {
	if !d.e.Dev {
		return
	}
	r.HandleFunc("/dev/reloadtemp", d.reloadTemplates).Methods("GET")
	r.HandleFunc("/dev/reloadconf", d.reloadConfiguration).Methods("GET")
}

// Dev mode route handlers
func (d *Dev) reloadTemplates(w http.ResponseWriter, r *http.Request) {
	if err := d.e.Views.Reload(); err != nil {
		w.Write([]byte("Error loading HTML templates: " + err.Error()))
		return
	}
	w.Write([]byte("Templates reloaded"))
}

func (d *Dev) reloadConfiguration(w http.ResponseWriter, r *http.Request) {
	if err := d.e.Config.Reload(); err != nil {
		w.Write([]byte("Error loading config: " + err.Error()))
		return
	}
	w.Write([]byte("Configuration reloaded"))
}
