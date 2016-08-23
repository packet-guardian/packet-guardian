// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package common

import (
	"errors"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
)

type Key int

const (
	SessionKey     Key = 0
	SessionUserKey Key = 1
	SessionEnvKey  Key = 2
)

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
		MaxAge: 0, // Expire on browser close
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

// Get and set Session to context moved to context files

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

// GetInt64 takes the same arguments as Get but def must be an int type.
func (s *Session) GetInt64(key interface{}, def ...int64) int64 {
	if v := s.Get(key); v != nil {
		return v.(int64)
	}
	if len(def) > 0 {
		return def[0]
	}
	return 0
}

func NewTestSession() *Session {
	return &Session{
		sessions.NewSession(&TestStore{}, "something"),
	}
}

type TestStore struct{}

func (t *TestStore) Get(r *http.Request, name string) (*sessions.Session, error) { return nil, nil }
func (t *TestStore) New(r *http.Request, name string) (*sessions.Session, error) { return nil, nil }
func (t *TestStore) Save(r *http.Request, w http.ResponseWriter, s *sessions.Session) error {
	return nil
}
