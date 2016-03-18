package main

import (
	"database/sql"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"

	log "github.com/dragonrider23/go-logger"
	"github.com/onesimus-systems/net-guardian/common"
	"github.com/onesimus-systems/net-guardian/dhcp"
)

var (
	configFile string
	dev        bool
)

func init() {
	flag.StringVar(&configFile, "config", "", "Configuration file path")
	flag.BoolVar(&dev, "devel", false, "Run in development mode")
}

func rootHandler(e *common.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := strings.Split(r.RemoteAddr, ":")[0]
		if dhcp.IsRegistered(e.DB, ip) {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		} else {
			http.Redirect(w, r, "/register", http.StatusTemporaryRedirect)
		}
	}
}

func main() {
	flag.Parse()
	if dev {
		fmt.Println(log.Magenta + "WARNING:" + log.Reset + " Net Guardian is running in DEVELOPMENT mode")
	}

	config := loadConfig("")
	sessStore := startSessionStore(config)
	db := connectDatabase(config)
	templates, err := template.ParseGlob("templates/*.tmpl")
	if err != nil {
		fmt.Println("Error loading HTML templates")
		os.Exit(1)
	}
	logger := log.New("app").Path("logs")
	if dev {
		logger.Verbose(3)
	}

	e := common.BuildEnvironment(logger, &sessionStore{sessStore}, db, config, templates, dev)

	r := mux.NewRouter()
	r.HandleFunc("/", rootHandler(e))
	r.PathPrefix("/public").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("./public/"))))
	r.HandleFunc("/register", dhcp.RegisterHTTPHandler(e)).Methods("GET")
	r.HandleFunc("/register/auto", dhcp.AutoRegisterHandler(e)).Methods("POST")

	startServer(r, config)
}

func connectDatabase(config *common.Config) *sql.DB {
	db, err := sql.Open("sqlite3", config.Core.DatabaseFile)
	if err != nil {
		fmt.Println("Error loading database file: ", config.Core.DatabaseFile)
		os.Exit(1)
	}
	return db
}

func startServer(router *mux.Router, config *common.Config) {
	bindAddr := ""
	bindPort := "8000"
	if config.Webserver.Address != "" {
		bindAddr = config.Webserver.Address
	}
	if config.Webserver.Port != 0 {
		bindPort = strconv.Itoa(config.Webserver.Port)
	}
	if bindAddr == "" {
		fmt.Printf("Now listening on *:%s\n", bindPort)
	} else {
		fmt.Printf("Now listening on %s:%s\n", bindAddr, bindPort)
	}
	http.ListenAndServe(bindAddr+":"+bindPort, router)
}
