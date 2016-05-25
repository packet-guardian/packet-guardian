// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/lfkeitel/verbose"
	"github.com/onesimus-systems/packet-guardian/src/common"

	"gopkg.in/tylerb/graceful.v1"
)

type Server struct {
	e         *common.Environment
	routes    http.Handler
	address   string
	httpPort  string
	httpsPort string
}

func NewServer(e *common.Environment, routes http.Handler) *Server {
	serv := &Server{
		e:       e,
		routes:  routes,
		address: e.Config.Webserver.Address,
	}

	serv.httpPort = strconv.Itoa(e.Config.Webserver.HttpPort)
	serv.httpsPort = strconv.Itoa(e.Config.Webserver.HttpsPort)
	return serv
}

func (s *Server) Run() {
	s.e.Log.Info("Starting web server...")
	if s.e.Config.Webserver.TLSCertFile == "" || s.e.Config.Webserver.TLSKeyFile == "" {
		s.startHttp()
		return
	}

	if s.e.Config.Webserver.RedirectHttpToHttps {
		go s.startRedirector()
	}
	s.startHttps()
}

func (s *Server) startRedirector() {
	s.e.Log.WithFields(verbose.Fields{
		"address": s.address,
		"port":    s.httpPort,
	}).Debug()
	srv := &graceful.Server{
		Timeout: 1 * time.Second,
		Server:  &http.Server{Addr: s.address + ":" + s.httpPort, Handler: http.HandlerFunc(s.redirectToHttps)},
	}
	if err := srv.ListenAndServe(); err != nil {
		s.e.Log.Fatal(err)
	}
}

func (s *Server) startHttp() {
	s.e.Log.WithFields(verbose.Fields{
		"address": s.address,
		"port":    s.httpPort,
	}).Debug()
	srv := &graceful.Server{
		Timeout: 5 * time.Second,
		Server:  &http.Server{Addr: s.address + ":" + s.httpPort, Handler: s.routes},
	}
	if err := srv.ListenAndServe(); err != nil {
		s.e.Log.Fatal(err)
	}
}

func (s *Server) startHttps() {
	s.e.Log.WithFields(verbose.Fields{
		"address": s.address,
		"port":    s.httpsPort,
	}).Debug()

	srv := &graceful.Server{
		Timeout: 5 * time.Second,
		Server:  &http.Server{Addr: s.address + ":" + s.httpsPort, Handler: s.routes},
	}
	if err := srv.ListenAndServeTLS(
		s.e.Config.Webserver.TLSCertFile,
		s.e.Config.Webserver.TLSKeyFile,
	); err != nil {
		s.e.Log.Fatal(err)
	}
}

func (s *Server) redirectToHttps(w http.ResponseWriter, r *http.Request) {
	// Lets not do a split if we don't need to
	if s.httpPort == "80" && s.httpsPort == "443" {
		http.Redirect(w, r, "https://"+r.Host+r.RequestURI, http.StatusMovedPermanently)
		return
	}

	host := strings.Split(r.Host, ":")[0]
	http.Redirect(w, r, "https://"+host+":"+s.httpsPort+r.RequestURI, http.StatusMovedPermanently)
}
