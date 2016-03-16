package main

import (
	"net/http"

	"github.com/onesimus-systems/net-guardian/common"
)

func init() {
	common.HTTPMux.HandleFunc("/overview", overviewHandler)
}

func overviewHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello"))
}
