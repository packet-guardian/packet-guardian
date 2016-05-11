// This file was generated with level_generator. DO NOT EDIT

package verbose

import (
	"fmt"
	"time"
)

// Entry represents a log message going through the system
type Entry struct {
	Level     LogLevel
	Timestamp time.Time
	Logger    *Logger
	Message   string
	Data      Fields
}

// NewEntry creates a new, empty Entry
func NewEntry(l *Logger) *Entry {
	return &Entry{
		Logger: l,
		Data:   make(Fields, 5),
	}
}

// WithField adds a single field to the Entry.
func (e *Entry) WithField(key string, value interface{}) *Entry {
	return e.WithFields(Fields{key: value})
}

// WithFields adds a map of fields to the Entry.
func (e *Entry) WithFields(fields Fields) *Entry {
	data := make(Fields, len(e.Data)+len(fields))
	for k, v := range e.Data {
		data[k] = v
	}
	for k, v := range fields {
		data[k] = v
	}
	return &Entry{Logger: e.Logger, Data: data}
}

// Log is the generic function to log a message with the handlers.
// All other logging functions are simply wrappers around this.
func (e *Entry) log(level LogLevel, msg string) {
	e.Logger.m.RLock()
	e.Level = level
	e.Message = msg
	e.Timestamp = time.Now()
	for _, h := range e.Logger.handlers {
		if h.Handles(level) {
			h.WriteLog(e)
		}
	}
	e.Logger.m.RUnlock()
}

// sprintlnn take from Logrus: github.com/Sirupsen/logrus entry.go

// Sprintlnn => Sprint no newline. This is to get the behavior of how
// fmt.Sprintln where spaces are always added between operands, regardless of
// their type. Instead of vendoring the Sprintln implementation to spare a
// string allocation, we do the simplest thing.
func (e *Entry) sprintlnn(args ...interface{}) string {
	msg := fmt.Sprintln(args...)
	return msg[:len(msg)-1]
}
