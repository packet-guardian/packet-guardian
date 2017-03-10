// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package common

import (
	"database/sql"
	"os"
	"os/signal"
	"syscall"

	"github.com/lfkeitel/verbose"
)

type EnvironmentEnv string

const (
	EnvTesting EnvironmentEnv = "testing"
	EnvProd    EnvironmentEnv = "production"
	EnvDev     EnvironmentEnv = "development"
)

type subscriber struct {
	c chan bool
}

// Environment holds "global" application information such as a database connection,
// logging, the config, sessions, etc.
type Environment struct {
	Sessions     *SessionStore
	DB           *DatabaseAccessor
	Config       *Config
	Views        *Views
	Env          EnvironmentEnv
	Log          *Logger
	shutdownSubs []*subscriber
	shutdownChan chan os.Signal
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
		stdout := verbose.NewStdoutHandler(true)
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

func (e *Environment) SubscribeShutdown() <-chan bool {
	e.shutdownWatcher() // Start the watcher

	sub := &subscriber{
		c: make(chan bool, 1),
	}

	e.shutdownSubs = append(e.shutdownSubs, sub)
	return sub.c
}

func (e *Environment) shutdownWatcher() {
	if e.shutdownChan != nil {
		return
	}

	e.shutdownChan = make(chan os.Signal, 1)
	signal.Notify(e.shutdownChan, os.Interrupt, syscall.SIGTERM)
	go func(env *Environment) {
		<-e.shutdownChan
		//e.Log.Notice("Calling shutdown subscribers")
		for _, sub := range e.shutdownSubs {
			//e.Log.Debugf("Calling shutdown subscriber %d", i)
			sub.c <- true
		}
	}(e)
}

type DatabaseAccessor struct {
	*sql.DB
	Driver string
}
