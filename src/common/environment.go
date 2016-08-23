// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package common

import (
	"database/sql"
	"errors"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite driver

	"github.com/lfkeitel/verbose"
)

type EnvironmentEnv string

const (
	EnvTesting EnvironmentEnv = "testing"
	EnvProd    EnvironmentEnv = "production"
	EnvDev     EnvironmentEnv = "development"
)

// Environment holds "global" application information such as a database connection,
// logging, the config, sessions, etc.
type Environment struct {
	Sessions *SessionStore
	DB       *DatabaseAccessor
	Config   *Config
	Views    *Views
	Env      EnvironmentEnv
	Log      *Logger
}

func NewEnvironment(t EnvironmentEnv) *Environment {
	return &Environment{Env: t}
}

func NewTestEnvironment() *Environment {
	e := &Environment{
		Config: NewEmptyConfig(),
		Log:    NewEmptyLogger(),
		Env:    EnvTesting,
	}
	// Disable automatic logging, manually configure if needed
	e.Log.c.Logging.Enabled = false
	if os.Getenv("PG_TEST_LOG") != "" {
		stdout := verbose.NewStdoutHandler()
		stdout.SetMinLevel(verbose.LogLevelDebug)
		e.Log.AddHandler("stdout", stdout)
	}
	return e
}

// Get and Set Environment to context, moved to context files

func (e *Environment) IsTesting() bool {
	return (e.Env == EnvTesting)
}

func (e *Environment) IsProd() bool {
	return (e.Env == EnvProd)
}

func (e *Environment) IsDev() bool {
	return (e.Env == EnvDev)
}

type DatabaseAccessor struct {
	*sql.DB
	Driver string
}

func NewDatabaseAccessor(config *Config) (*DatabaseAccessor, error) {
	var err error
	da := &DatabaseAccessor{}

	if config.Database.Type == "sqlite" {
		if !FileExists(config.Database.Address) {
			return nil, errors.New("SQLite database file doesn't exist")
		}
		da.DB, err = sql.Open("sqlite3", config.Database.Address)
		da.Driver = "sqlite3"
	} else {
		return nil, errors.New("Unsupported database type " + config.Database.Type)
	}

	if err != nil {
		return nil, err
	}
	da.setupDatabase()
	return da, nil
}

func (da *DatabaseAccessor) setupDatabase() error {
	var err error
	switch da.Driver {
	case "sqlite3":
		_, err = da.Exec("PRAGMA foreign_keys = ON")
	}
	return err
}

var logLevels = map[string]verbose.LogLevel{
	"debug":     verbose.LogLevelDebug,
	"info":      verbose.LogLevelInfo,
	"notice":    verbose.LogLevelNotice,
	"warning":   verbose.LogLevelWarning,
	"error":     verbose.LogLevelError,
	"critical":  verbose.LogLevelCritical,
	"alert":     verbose.LogLevelAlert,
	"emergency": verbose.LogLevelEmergency,
	"fatal":     verbose.LogLevelFatal,
}

type Logger struct {
	*verbose.Logger
	c      *Config
	timers map[string]time.Time
}

func NewEmptyLogger() *Logger {
	return &Logger{
		Logger: verbose.New("null"),
		timers: make(map[string]time.Time),
		c:      &Config{},
	}
}

func NewLogger(c *Config, name string) *Logger {
	logger := verbose.New(name)
	if !c.Logging.Enabled {
		return &Logger{
			Logger: logger,
		}
	}
	sh := verbose.NewStdoutHandler()
	fh, _ := verbose.NewFileHandler(c.Logging.Path)
	logger.AddHandler("stdout", sh)
	logger.AddHandler("file", fh)

	if level, ok := logLevels[strings.ToLower(c.Logging.Level)]; ok {
		sh.SetMinLevel(level)
		fh.SetMinLevel(level)
	}
	// The verbose package sets the default max to Emergancy
	sh.SetMaxLevel(verbose.LogLevelFatal)
	fh.SetMaxLevel(verbose.LogLevelFatal)
	return &Logger{
		Logger: logger,
		c:      c,
	}
}

// GetLogger returns a new Logger based on its parent but with a new name
// This can be used to separate logs from different sub-systems.
func (l *Logger) GetLogger(name string) *Logger {
	return NewLogger(l.c, name)
}

func (l *Logger) StartTimer(name string) {
	l.timers[name] = time.Now()
}

func (l *Logger) StopTimer(name string) {
	t, ok := l.timers[name]
	if !ok {
		return
	}
	l.Debugf("Timer %s duration %s", name, time.Since(t).String())
	delete(l.timers, name)
}
