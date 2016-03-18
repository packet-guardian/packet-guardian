package common

import (
	"database/sql"
	"html/template"
	"net/http"

	log "github.com/dragonrider23/go-logger"
)

// Environment holds "global" application information such as a database connection,
// logging, the config, sessions, etc.
type Environment struct {
	Sessions  SessionStore
	DB        *sql.DB
	Config    *Config
	Templates *template.Template
	Dev       bool
	Log       *log.Logger
}

// A SessionStore is a wrapper around gorilla's session store type
type SessionStore interface {
	GetSession(r *http.Request, name string) Session
}

// A Session is a wrapper around gorilla's session type
type Session interface {
	Get(key interface{}, def ...interface{}) interface{}
	Set(key, val interface{})
	GetBool(key interface{}, def ...bool) bool
	GetString(key interface{}, def ...string) string
	GetInt(key interface{}, def ...int) int
}

// BuildEnvironment takes the needed types and returns an Environment
func BuildEnvironment(l *log.Logger, s SessionStore, d *sql.DB, c *Config, t *template.Template, devel bool) *Environment {
	e := &Environment{
		Log:       l,
		Sessions:  s,
		DB:        d,
		Config:    c,
		Templates: t,
		Dev:       devel,
	}
	return e
}

// BuildEmptyEnvironment returns an empty environment for testing purposes
func BuildEmptyEnvironment() *Environment {
	logger := log.New("null")
	logger.NoFile()
	logger.NoStdout()
	return BuildEnvironment(logger, nil, nil, nil, nil, false)
}
