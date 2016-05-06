package server

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/dragonrider23/verbose"
	"github.com/onesimus-systems/packet-guardian/src/common"
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
	s.e.Log.Critical(http.ListenAndServe(
		s.address+":"+s.httpPort,
		http.HandlerFunc(s.redirectToHttps),
	))
}

func (s *Server) startHttp() {
	s.e.Log.WithFields(verbose.Fields{
		"address": s.address,
		"port":    s.httpPort,
	}).Debug()
	s.e.Log.Fatal(http.ListenAndServe(
		s.address+":"+s.httpPort,
		s.routes,
	))
}

func (s *Server) startHttps() {
	s.e.Log.WithFields(verbose.Fields{
		"address": s.address,
		"port":    s.httpsPort,
	}).Debug()
	s.e.Log.Fatal(http.ListenAndServeTLS(
		s.address+":"+s.httpsPort,
		s.e.Config.Webserver.TLSCertFile,
		s.e.Config.Webserver.TLSKeyFile,
		s.routes,
	))
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
