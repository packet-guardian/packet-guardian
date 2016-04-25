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

	if e.Config.Webserver.HttpPort == 0 {
		serv.httpPort = "8080"
	} else {
		serv.httpPort = strconv.Itoa(e.Config.Webserver.HttpPort)
	}

	if e.Config.Webserver.HttpsPort == 0 {
		serv.httpsPort = "1443"
	} else {
		serv.httpsPort = strconv.Itoa(e.Config.Webserver.HttpsPort)
	}
	return serv
}

func (s *Server) Run() {
	if s.e.Config.Webserver.TLSCertFile != "" && s.e.Config.Webserver.TLSKeyFile != "" {
		if s.e.Config.Webserver.RedirectHttpToHttps {
			go func() {
				s.e.Log.Infof("Now listening on %s:%s", s.address, s.httpPort)
				http.ListenAndServe(s.address+":"+s.httpPort, http.HandlerFunc(s.redirectToHttps))
			}()
		}
		s.startHttps()
	} else {
		s.startHttp()
	}
}

func (s *Server) startHttp() {
	s.e.Log.Infof("Now listening on %s:%s", s.address, s.httpPort)
	http.ListenAndServe(s.address+":"+s.httpPort, s.routes)
}

func (s *Server) startHttps() {
	s.e.Log.WithFields(verbose.Fields{
		"address": s.address,
		"port":    s.httpsPort,
	}).Info("Now listening on TLS")
	s.e.Log.Info("Starting server with TLS certificates")
	http.ListenAndServeTLS(
		s.address+":"+s.httpsPort,
		s.e.Config.Webserver.TLSCertFile,
		s.e.Config.Webserver.TLSKeyFile,
		s.routes,
	)
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
