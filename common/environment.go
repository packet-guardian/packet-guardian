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
	DHCP      DHCPHostWriter
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
	Save(r *http.Request, w http.ResponseWriter) error
}

// A DHCPHostWriter can write a new DHCPd hosts file
type DHCPHostWriter interface {
	WriteHostFile()
}
