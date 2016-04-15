package server

import (
	"net/http"
	"strconv"

	"github.com/onesimus-systems/packet-guardian/src/common"
)

type Server struct {
	e      *common.Environment
	routes http.Handler
}

func NewServer(e *common.Environment, routes http.Handler) *Server {
	return &Server{
		e:      e,
		routes: routes,
	}
}

func (s *Server) Run() {
	bindAddr := ""
	bindPort := "8000"
	if s.e.Config.Webserver.Address != "" {
		bindAddr = s.e.Config.Webserver.Address
	}
	if s.e.Config.Webserver.Port != 0 {
		bindPort = strconv.Itoa(s.e.Config.Webserver.Port)
	}
	if bindAddr == "" {
		s.e.Log.Infof("Now listening on *:%s", bindPort)
	} else {
		s.e.Log.Infof("Now listening on %s:%s", bindAddr, bindPort)
	}

	if s.e.Config.Webserver.TLSCertFile != "" && s.e.Config.Webserver.TLSKeyFile != "" {
		s.e.Log.Info("Starting server with TLS certificates")
		http.ListenAndServeTLS(
			bindAddr+":"+bindPort,
			s.e.Config.Webserver.TLSCertFile,
			s.e.Config.Webserver.TLSKeyFile,
			s.routes,
		)
	} else {
		http.ListenAndServe(bindAddr+":"+bindPort, s.routes)
	}
}
