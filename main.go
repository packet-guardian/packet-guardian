package main

import (
	"database/sql"
	"errors"
	"flag"
	"html/template"
	"net/http"
	"strconv"

	_ "github.com/mattn/go-sqlite3"

	log "github.com/dragonrider23/go-logger"
	"github.com/onesimus-systems/packet-guardian/common"
)

var (
	configFile string
	dev        bool
)

func init() {
	flag.StringVar(&configFile, "config", "", "Configuration file path")
	flag.BoolVar(&dev, "dev", false, "Run in development mode")
}

func main() {
	// Parse CLI flags
	flag.Parse()

	// Get application-wide resources
	logger := log.New("app").Path("logs")
	if dev {
		logger.Verbose(3)
		logger.Info("Packet Guardian running in DEVELOPMENT mode")
	}
	config, err := loadConfig(configFile)
	if err != nil {
		logger.Fatalf("Error loading configuration: %s", err.Error())
	}
	logger.Infof("Configuration loaded from %s", config.SourceFile)
	sessStore, err := startSessionStore(config)
	if err != nil {
		logger.Fatalf("Error loading session store: %s", err.Error())
	}
	db, err := connectDatabase(config)
	if err != nil {
		logger.Fatalf("Error loading database: %s", err.Error())
	}
	templates, err := parseTemplates("templates/*.tmpl")
	if err != nil {
		logger.Fatalf("Error loading HTML templates: %s", err.Error())
	}

	// Create an environment
	e := &common.Environment{
		Log:       logger,
		Sessions:  &sessionStore{sessStore},
		DB:        db,
		Config:    config,
		Templates: templates,
		Dev:       dev,
	}

	// Let's begin!
	startServer(makeRoutes(e), e)
}

func parseTemplates(pattern string) (tmpl *template.Template, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
			tmpl = nil
		}
	}()

	tmpl = template.Must(template.New("").Funcs(template.FuncMap{
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, errors.New("invalid dict call")
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, errors.New("dict keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
		"list": func(values ...interface{}) ([]interface{}, error) {
			return values, nil
		},
	}).ParseGlob(pattern))
	return
}

func connectDatabase(config *common.Config) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", config.Core.DatabaseFile)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func startServer(router http.Handler, e *common.Environment) {
	bindAddr := ""
	bindPort := "8000"
	if e.Config.Webserver.Address != "" {
		bindAddr = e.Config.Webserver.Address
	}
	if e.Config.Webserver.Port != 0 {
		bindPort = strconv.Itoa(e.Config.Webserver.Port)
	}
	if bindAddr == "" {
		e.Log.Infof("Now listening on *:%s", bindPort)
	} else {
		e.Log.Infof("Now listening on %s:%s", bindAddr, bindPort)
	}

	if e.Config.Webserver.TLSCertFile != "" && e.Config.Webserver.TLSKeyFile != "" {
		e.Log.Info("Starting server with TLS certificates")
		http.ListenAndServeTLS(
			bindAddr+":"+bindPort,
			e.Config.Webserver.TLSCertFile,
			e.Config.Webserver.TLSKeyFile,
			router,
		)
	} else {
		http.ListenAndServe(bindAddr+":"+bindPort, router)
	}
}
