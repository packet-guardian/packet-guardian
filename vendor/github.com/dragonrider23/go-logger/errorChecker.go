// Copyright 2014 Lee Keitel. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package logger

// CheckError will check error e and, if
// not nil, will write the error to the
// logger l as type Error. CheckError
// will return a bool value true if there
// was an error or false if e is nil.
// CheckError will use the logger with an
// empty name "" by default if one isn't given.
func CheckError(e error, l *Logger) bool {
	if e == nil {
		return false
	}

	if l == nil {
		l = Get("")
	}

	l.Error(e.Error())
	return true
}
