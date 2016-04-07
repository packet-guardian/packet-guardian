package main

import (
	"errors"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/onesimus-systems/packet-guardian/common"
)

// StartSessionStore opens the FilesystemStore for web sessions
func startSessionStore(config *common.Config) (*sessions.FilesystemStore, error) {
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

	store := sessions.NewFilesystemStore(sessDir, sessKeyPair...)

	store.Options = &sessions.Options{
		Path:   "/",
		MaxAge: 3600 * 8, // 8 hours
	}
	return store, nil
}

type sessionStore struct {
	*sessions.FilesystemStore
}

// GetSession takes the same arguments as gorilla/sessions Get and returns a
// Session object. Note this is the common.Session type NOT gorilla/sessions.Session
func (s *sessionStore) GetSession(r *http.Request, name string) common.Session {
	sess, _ := s.Get(r, name)
	return &session{sess}
}

// Session is a wrapper around Gorilla sessions to provide access methods
type session struct {
	*sessions.Session
}

func (s *session) Delete(r *http.Request, w http.ResponseWriter) error {
	s.Options.MaxAge = -1
	return s.Save(r, w)
}

// Get a value from the session object
func (s *session) Get(key interface{}, def ...interface{}) interface{} {
	if val, ok := s.Values[key]; ok {
		return val
	}
	if len(def) > 0 {
		return def[0]
	}
	return nil
}

// Set a value to the session object
func (s *session) Set(key, val interface{}) {
	s.Values[key] = val
}

// GetBool takes the same arguments as Get but def must be a bool type.
func (s *session) GetBool(key interface{}, def ...bool) bool {
	if v := s.Get(key); v != nil {
		return v.(bool)
	}
	if len(def) > 0 {
		return def[0]
	}
	return false
}

// GetString takes the same arguments as Get but def must be a string type.
func (s *session) GetString(key interface{}, def ...string) string {
	if v := s.Get(key); v != nil {
		return v.(string)
	}
	if len(def) > 0 {
		return def[0]
	}
	return ""
}

// GetInt takes the same arguments as Get but def must be an int type.
func (s *session) GetInt(key interface{}, def ...int) int {
	if v := s.Get(key); v != nil {
		return v.(int)
	}
	if len(def) > 0 {
		return def[0]
	}
	return 0
}
