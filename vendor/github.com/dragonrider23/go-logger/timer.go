// Copyright 2014 Lee Keitel. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package logger

import (
	"strings"
	"time"
)

const timeReplacementString string = "{time}"

// timer holds the start time for a log timer.
// Each logger can only have one timer
type timer struct {
	start   time.Time
	running bool
}

// StartTimer associates a timer with logger l
func (l *Logger) StartTimer() {
	l.t = timer{
		start:   time.Now(),
		running: true,
	}
	return
}

// StopTimer determines the time elapsed since logger
// l's timer started and issues an Info level log using string s.
// The string "{time}" will be replaced with the elapsed time in the log message.
// If s is empty, no log will be written. Returns the elapsed time as a string.
func (l *Logger) StopTimer(s string) string {
	if !l.t.running {
		return ""
	}

	elapsed := time.Since(l.t.start).String()
	if s != "" {
		s = strings.Replace(s, timeReplacementString, elapsed, -1)
		l.Info(s)
	}
	l.t.running = false
	return elapsed
}
