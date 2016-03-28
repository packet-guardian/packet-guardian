package main

import (
	"database/sql"
	"errors"
	"flag"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"

	log "github.com/dragonrider23/go-logger"
	"github.com/onesimus-systems/packet-guardian/auth"
	"github.com/onesimus-systems/packet-guardian/common"
	"github.com/onesimus-systems/packet-guardian/dhcp"
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
		if reg, _ := dhcp.IsRegistered(e.DB, ip); reg {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		} else {
			http.Redirect(w, r, "/register", http.StatusTemporaryRedirect)
		}
	}
}

func fileHandler(e *common.Environment, template string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			SiteTitle   string
			CompanyName string
		}{
			SiteTitle:   e.Config.Core.SiteTitle,
			CompanyName: e.Config.Core.SiteCompanyName,
		}
		e.Templates.ExecuteTemplate(w, template, data)
	}
}

func reloadTemplates(e *common.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		templates, err := parseTemplates("templates/*.tmpl")
		if err != nil {
			w.Write([]byte("Error loading HTML templates: " + err.Error()))
			return
		}
		e.Templates = templates
		w.Write([]byte("Templates reloaded"))
	}
}

func reloadConfiguration(e *common.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		config, err := loadConfig("")
		if err != nil {
			w.Write([]byte("Error loading config: " + err.Error()))
			return
		}
		e.Config = config
		w.Write([]byte("Configuration reloaded"))
	}
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
	config, err := loadConfig("")
	if err != nil {
		logger.Fatalf("Error loading configuration: %s", err.Error())
	}
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
	c, q := dhcp.StartHostWriteService(db, config.DHCP.HostsFile)
	e := &common.Environment{
		Log:       logger,
		Sessions:  &sessionStore{sessStore},
		DB:        db,
		Config:    config,
		Templates: templates,
		Dev:       dev,
		DHCP: &dhcpHostFile{
			write: c,
			quit:  q,
		},
	}

	// Register HTTP handlers
	r := mux.NewRouter()
	r.HandleFunc("/", rootHandler(e))
	r.PathPrefix("/public").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("./public/"))))

	r.HandleFunc("/register", fileHandler(e, "register")).Methods("GET")
	r.HandleFunc("/register", dhcp.AutoRegisterHandler(e)).Methods("POST")

	r.HandleFunc("/login", fileHandler(e, "login")).Methods("GET")
	r.HandleFunc("/login", auth.LoginHandler(e)).Methods("POST")

	if dev {
		r.HandleFunc("/dev/reloadtemp", reloadTemplates(e)).Methods("GET")
		r.HandleFunc("/dev/reloadconf", reloadConfiguration(e)).Methods("GET")
	}

	// Let's begin!
	startServer(r, e)
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
	http.ListenAndServe(bindAddr+":"+bindPort, router)
}

type dhcpHostFile struct {
	write chan bool
	quit  chan bool
}

func (d *dhcpHostFile) WriteHostFile() {
	select {
	case d.write <- true:
	default:
	}
}
