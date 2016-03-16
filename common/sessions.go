package common

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
)

// Store is the web session store
var store *sessions.FilesystemStore

// StartSessionStore opens the FilesystemStore for web sessions
func StartSessionStore() {
	if Config.Webserver.SessionsDir == "" {
		Config.Webserver.SessionsDir = "sessions"
	}
	if Config.Webserver.SessionsAuthKey == "" {
		fmt.Println("No session authentication key given in configuration")
		os.Exit(1)
	}

	err := os.MkdirAll(Config.Webserver.SessionsDir, 0700)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	sessDir := Config.Webserver.SessionsDir
	sessKeyPair := make([][]byte, 1)
	sessKeyPair[0] = []byte(Config.Webserver.SessionsAuthKey)
	if Config.Webserver.SessionsEncryptKey != "" {
		sessKeyPair = append(sessKeyPair, []byte(Config.Webserver.SessionsEncryptKey))
	}

	store = sessions.NewFilesystemStore(sessDir, sessKeyPair...)

	store.Options = &sessions.Options{
		Path:   "/",
		MaxAge: 3600 * 8, // 8 hours
	}
}

// Session is a wrapper around Gorilla sessions to provide access methods
type Session struct {
	*sessions.Session
}

// GetSession takes the same arguments as gorilla/sessions Get and returns a
// Session object. Note this is the common.Session type NOT gorilla/sessions.Session
func GetSession(r *http.Request, name string) *Session {
	session, _ := store.Get(r, name)
	return &Session{session}
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
