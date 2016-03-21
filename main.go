package main

import (
	"database/sql"
	"errors"
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
		e.Templates.ExecuteTemplate(w, template, nil)
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

func main() {
	// Parse CLI flags
	flag.Parse()
	if dev {
		fmt.Println(log.Magenta + "WARNING:" + log.Reset + " Net Guardian is running in DEVELOPMENT mode")
	}

	// Get application-wide resources
	logger := log.New("app").Path("logs")
	config := loadConfig("")
	sessStore := startSessionStore(config)
	db := connectDatabase(config)
	templates, err := parseTemplates("templates/*.tmpl")
	if err != nil {
		fmt.Printf("Error loading HTML templates: %s", err.Error())
		os.Exit(1)
	}
	if dev {
		logger.Verbose(3)
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
		r.HandleFunc("/dev/reload", reloadTemplates(e)).Methods("GET")
	}

	// Let's begin!
	startServer(r, config)
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

func connectDatabase(config *common.Config) *sql.DB {
	db, err := sql.Open("sqlite3", config.Core.DatabaseFile)
	if err != nil {
		fmt.Println("Error loading database file: ", config.Core.DatabaseFile)
		os.Exit(1)
	}
	return db
}

func startServer(router http.Handler, config *common.Config) {
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
