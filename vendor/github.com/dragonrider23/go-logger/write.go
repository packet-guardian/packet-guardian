// Copyright 2014 Lee Keitel. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package logger

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

// Wrapper function to call both writeToStdout and writeToFile
func (l *Logger) writeAll(e, s, c string) (n int, err error) {
	l.writeToStdout(e, s, c)
	n, err = l.writeToFile(e, s)
	return
}

// Write log text to stdout
func (l *Logger) writeToStdout(e, s, c string) {
	if !l.stdout {
		return
	}

	// Check verbosity
	log := false
	if l.verbosity == 3 { // All levels including custom
		log = true
	}
	if l.verbosity == 2 && e != "Info" { // Warning, Error, Fatal
		log = true
	}
	if l.verbosity == 1 && (e != "Info" && e != "Warning") { // Error, Fatal
		log = true
	}
	if l.verbosity == 0 && e == "Fatal" { // Fatal
		log = true
	}
	if !log {
		return
	}

	now := time.Now().Format(l.tlayout)
	fmt.Printf("%s%s: %s%s: %s%s\n", Grey, now, c, strings.ToUpper(e), Reset, s)
	return
}

// Write log text specific path with filename [l.name]-[e].log and path l.path
func (l *Logger) writeToFile(e, s string) (n int, err error) {
	if !l.file {
		return 0, errors.New("Write to file is disabled for this logger")
	}
	if err = checkPath(l.path); err != nil {
		return 0, err
	}

	// Prepare time stamp
	t := ""
	if !l.raw {
		t = time.Now().Format(l.tlayout) + ": "
	}

	// Prepare filename
	var loggerName string
	if l.name == "" {
		loggerName = ""
	} else {
		loggerName = strings.ToLower(l.name) + "-"
	}
	fileName := l.path + loggerName + strings.ToLower(e) + ".log"
	errorStr := t + s + "\n"

	// Open and write to logfile
	saveFile, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println(err.Error())
	}
	defer saveFile.Close()

	n, err = saveFile.WriteString(errorStr)
	if err != nil {
		fmt.Println(err.Error())
	}
	return
}

// Checks file path to make sure it's available and if not creates it
func checkPath(p string) error {
	_, err := os.Stat(p)
	if err == nil {
		return nil
	}

	if os.IsNotExist(err) {
		if err = os.Mkdir(p, 0755); err != nil {
			return errors.New("ERROR: Logger - Couldn't create logs folder")
		}
		return nil
	}
	return errors.New("ERROR: Logger - Unknown file error")
}
