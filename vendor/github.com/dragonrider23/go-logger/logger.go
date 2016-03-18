// Copyright 2014 Lee Keitel. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// The logger package allows for organized, simplified, custom
// logging for any application. Logger makes creating logs easy.
// Create multiple loggers for specialized purposes each with their
// own specific settings such as verbosity, log file location, and
// whether it's show on stdout, written to file, or only one of them.
// Logger comes with 4 pre-made logging levels ready for use. There's
// also a generic Log() function that allows for custom log types.
//
// Logger also has timers that can be attached to logs. This makes it
// easy to track how long a function or request takes to run. Logger
// also comes with a wrapper function to check for errors in an application.
// Instead of having to write out the log to file code for every err if
// statement, CheckError() will take the error, check if one exists and
// then write it to the logger of choice. It will also return boolean
// to tell the calling function if an error happened or not so any custom
// actions can be performed if needed.
package logger

import (
	"fmt"
	"os"
)

// Colors that can be used for Log().
const (
	Reset   = "\033[0m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"
	Grey    = "\x1B[90m"
)

var verbosity int = 2

// Set global verbosity for stdout.
func Verbose(v int) {
	verbosity = checkVerboseLevel(v)
	return
}

// Gets current global verbosity.
func GetVerboseLevel() int {
	return verbosity
}

// Validate verbosity level.
func checkVerboseLevel(v int) int {
	if v < 0 {
		v = 0
	} else if v > 3 {
		v = 3
	}
	return v
}

// Type logger is the struct returned and used for logging.
// The user can set its properties using the associated functions.
type Logger struct {
	name, path, tlayout string
	stdout, file, raw   bool
	verbosity           int
	t                   timer
}

// loggers is the collection of all logger types used.
// A logger can be called by their names instead of
// passing around a *Logger.
var loggers map[string]*Logger

// New returns a pointer to a new logger struct named n.
func New(n string) *Logger {
	// Initialize loggers
	if loggers == nil {
		loggers = make(map[string]*Logger)
	}
	// If logger with n name is already created, return it
	if loggers[n] != nil {
		return loggers[n]
	}
	// Create new logger
	newLogger := &Logger{
		name:      n,
		stdout:    true,
		file:      true,
		raw:       false,
		verbosity: verbosity,
		path:      "logs/",
		tlayout:   "2006-01-02 15:04:05 MST",
	}
	// Add to loggers
	loggers[n] = newLogger
	return newLogger
}

// Get retrives the logger with name n.
func Get(n string) *Logger {
	log := loggers[n]
	// If logger doesn't exist, create it
	if loggers[n] == nil {
		log = New(n)
	}
	return log
}

// Set verbose level of logger.
func (l *Logger) Verbose(v int) *Logger {
	l.verbosity = checkVerboseLevel(v)
	return l
}

// NoStdout disables the logger from going to stdout.
func (l *Logger) NoStdout() *Logger {
	l.stdout = false
	return l
}

// Stdout enables the logger going to stdout.
func (l *Logger) Stdout() *Logger {
	l.stdout = true
	return l
}

// NoFile disables the logger from writting to a file.
func (l *Logger) NoFile() *Logger {
	l.file = false
	return l
}

// File enables the logger writting to a file.
func (l *Logger) File() *Logger {
	l.file = true
	return l
}

// Raw tells the logger writer to not pre-include the date.
// Allows for completely custom log text.
func (l *Logger) Raw() *Logger {
	l.raw = true
	return l
}

// Path sets the filepath for the log files.
func (l *Logger) Path(p string) *Logger {
	if p[len(p)-1] != '/' {
		p += "/"
	}
	l.path = p
	return l
}

// TimeCode sets the time layout used in the logs.
func (l *Logger) TimeCode(t string) *Logger {
	l.tlayout = t
	return l
}

// Remove logger l from the loggers map.
func (l *Logger) Close() {
	delete(loggers, l.name)
	return
}

// Wrapper for Log("Info", ...). Shows blue in stdout.
func (l *Logger) Info(format string) {
	l.Log("Info", Cyan, format)
	return
}

// Wrapper for formatted Log("Info", ...). Shows blue in stdout.
func (l *Logger) Infof(format string, v ...interface{}) {
	l.Log("Info", Cyan, format, v...)
	return
}

// Wrapper for Log("Warning", ...). Shows magenta in stdout.
func (l *Logger) Warning(format string) {
	l.Log("Warning", Magenta, format)
	return
}

// Wrapper for formatted Log("Warning", ...). Shows magenta in stdout.
func (l *Logger) Warningf(format string, v ...interface{}) {
	l.Log("Warning", Magenta, format, v...)
	return
}

// Wrapper for Log("Error", ...). Shows red in stdout.
func (l *Logger) Error(format string) {
	l.Log("Error", Red, format)
	return
}

// Wrapper for formatted Log("Error", ...). Shows red in stdout.
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.Log("Error", Red, format, v...)
	return
}

// Wrapper for Log("Fatal", ...). Shows red in stdout.
// Exits application with os.Exit(1).
func (l *Logger) Fatal(format string) {
	l.Log("Fatal", Red, format)
	os.Exit(1)
	return
}

// Wrapper for formatted Log("Fatal", ...). Shows red in stdout.
// Exits application with os.Exit(1).
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.Log("Fatal", Red, format, v...)
	os.Exit(1)
	return
}

// Log is the core function that will write a log text to file and stdout.
// The log will be of eType type (used for the filename of the log). In
// stdout it will be colored color (see const list). The text will use format
// to Printf v interfaces.
func (l *Logger) Log(eType, color, format string, v ...interface{}) {
	if color == "" {
		color = White
	}
	l.writeAll(eType, fmt.Sprintf(format, v...), color)
	return
}
