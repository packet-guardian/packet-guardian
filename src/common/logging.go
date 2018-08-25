// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package common

import (
	"fmt"
	"strings"
	"time"

	"github.com/lfkeitel/verbose/v4"
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

	setupStdOutLogging(logger, c.Logging.Level)
	setupFileLogging(logger, c.Logging.Level, c.Logging.Path)

	return &Logger{
		Logger: logger,
		c:      c,
	}
}

func setupStdOutLogging(logger *verbose.Logger, level string) {
	sh := verbose.NewStdoutHandler(true)
	setMinimumLoggingLevel(sh, level)
	logger.AddHandler("stdout", sh)
}

func setupFileLogging(logger *verbose.Logger, level, path string) {
	if path == "" {
		return
	}

	fh, err := verbose.NewFileHandler(path)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fh.SetFormatter(verbose.NewJSONFormatter())
	setMinimumLoggingLevel(fh, level)
	logger.AddHandler("file", fh)
}

func setMinimumLoggingLevel(logger verbose.Handler, level string) {
	if level, ok := logLevels[strings.ToLower(level)]; ok {
		logger.SetMinLevel(level)
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
