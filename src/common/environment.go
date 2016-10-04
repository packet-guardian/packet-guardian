// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package common

import (
	"os"

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
