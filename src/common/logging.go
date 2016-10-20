// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package common

import (
	"strings"
	"time"

	"github.com/lfkeitel/verbose"
)

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

var SystemLogger *Logger

func init() {
	SystemLogger = NewEmptyLogger()
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
	sh := verbose.NewStdoutHandler(true)
	fh, _ := verbose.NewFileHandler(c.Logging.Path)
	logger.AddHandler("stdout", sh)
	logger.AddHandler("file", fh)

	if level, ok := logLevels[strings.ToLower(c.Logging.Level)]; ok {
		sh.SetMinLevel(level)
		fh.SetMinLevel(level)
	}
	fh.SetFormatter(&verbose.JSONFormatter{})
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
