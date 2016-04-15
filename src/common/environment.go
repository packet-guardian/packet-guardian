package common

import (
	"database/sql"
	"errors"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3" // SQLite driver

	log "github.com/dragonrider23/go-logger"
	"github.com/gorilla/sessions"
)

// Environment holds "global" application information such as a database connection,
// logging, the config, sessions, etc.
type Environment struct {
	Sessions *SessionStore
	DB       *DatabaseAccessor
	Config   *Config
	Views    *Views
	Dev      bool
	Log      *Logger
}

func NewEnvironment(dev bool) *Environment {
	return &Environment{Dev: dev}
}

type DatabaseAccessor struct {
	*sql.DB
}

func NewDatabaseAccessor(file string) (*DatabaseAccessor, error) {
	db, err := sql.Open("sqlite3", file)
	if err != nil {
		return nil, err
	}
	return &DatabaseAccessor{db}, nil
}

type Views struct {
	source string
	t      *template.Template
	e      *Environment
}

func NewViews(e *Environment, basepath string) (v *Views, err error) {
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
		}
	}()

	tmpl := template.New("").Funcs(template.FuncMap{
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
		"config": func() *Config {
			return e.Config
		},
	})

	filepath.Walk(basepath, func(path string, info os.FileInfo, err1 error) error {
		if strings.HasSuffix(path, ".tmpl") {
			if _, err := tmpl.ParseFiles(path); err != nil {
				panic(err)
			}
		}
		return nil
	})
	v = &Views{
		source: basepath,
		t:      tmpl,
		e:      e,
	}
	return v, nil
}

func (v *Views) NewView(view string) *View {
	return &View{
		name: view,
		t:    v.t,
		e:    v.e,
	}
}

func (v *Views) Reload() error {
	views, err := NewViews(v.e, v.source)
	if err != nil {
		return err
	}
	v.t = views.t
	return nil
}

type View struct {
	name string
	t    *template.Template
	e    *Environment
}

func (v *View) Render(w io.Writer, data interface{}) error {
	return v.t.ExecuteTemplate(w, v.name, data)
}

type SessionStore struct {
	*sessions.FilesystemStore
	sessionName string
}

func NewSessionStore(config *Config) (*SessionStore, error) {
	if config.Webserver.SessionsDir == "" {
		config.Webserver.SessionsDir = "sessions"
	}
	if config.Webserver.SessionsAuthKey == "" {
		return nil, errors.New("No session authentication key given in configuration")
	}

	err := os.MkdirAll(config.Webserver.SessionsDir, 0700)
	if err != nil {
		return nil, err
	}

	sessDir := config.Webserver.SessionsDir
	sessKeyPair := make([][]byte, 1)
	sessKeyPair[0] = []byte(config.Webserver.SessionsAuthKey)
	if config.Webserver.SessionsEncryptKey != "" {
		sessKeyPair = append(sessKeyPair, []byte(config.Webserver.SessionsEncryptKey))
	}

	store := &SessionStore{
		FilesystemStore: sessions.NewFilesystemStore(sessDir, sessKeyPair...),
		sessionName:     config.Webserver.SessionName,
	}

	store.Options = &sessions.Options{
		Path:   "/",
		MaxAge: 3600 * 8, // 8 hours
	}
	return store, nil
}

// GetSession returns a session based on the http request.
func (s *SessionStore) GetSession(r *http.Request) *Session {
	sess, _ := s.Get(r, s.sessionName)
	return &Session{sess}
}

// Session is a wrapper around Gorilla sessions to provide access methods
type Session struct {
	*sessions.Session
}

func (s *Session) Delete(r *http.Request, w http.ResponseWriter) error {
	s.Options.MaxAge = -1
	return s.Save(r, w)
}

// Get a value from the session object
func (s *Session) Get(key interface{}, def ...interface{}) interface{} {
	if val, ok := s.Values[key]; ok {
		return val
	}
	if len(def) > 0 {
		return def[0]
	}
	return nil
}

// Set a value to the session object
func (s *Session) Set(key, val interface{}) {
	s.Values[key] = val
}

// GetBool takes the same arguments as Get but def must be a bool type.
func (s *Session) GetBool(key interface{}, def ...bool) bool {
	if v := s.Get(key); v != nil {
		return v.(bool)
	}
	if len(def) > 0 {
		return def[0]
	}
	return false
}

// GetString takes the same arguments as Get but def must be a string type.
func (s *Session) GetString(key interface{}, def ...string) string {
	if v := s.Get(key); v != nil {
		return v.(string)
	}
	if len(def) > 0 {
		return def[0]
	}
	return ""
}

// GetInt takes the same arguments as Get but def must be an int type.
func (s *Session) GetInt(key interface{}, def ...int) int {
	if v := s.Get(key); v != nil {
		return v.(int)
	}
	if len(def) > 0 {
		return def[0]
	}
	return 0
}

type Logger struct {
	*log.Logger
}

func NewLogger(logDir string, dev bool) *Logger {
	logger := log.New("app").Path(logDir)
	if dev {
		logger.Verbose(3)
		logger.Info("Packet Guardian running in DEVELOPMENT mode")
	}
	return &Logger{logger}
}
