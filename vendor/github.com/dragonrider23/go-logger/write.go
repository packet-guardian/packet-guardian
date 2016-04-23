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

var verboseLevels = map[string]int{
	"Info":    3,
	"Warning": 2,
	"Error":   1,
	"Fatal":   0,
}

// Wrapper function to call both writeToStdout and writeToFile
func (l *Logger) writeAll(e, s, c string) error {
	l.writeToStdout(e, s, c)
	err := l.writeToFile(e, s)
	return err
}

// Write log text to stdout
func (l *Logger) writeToStdout(e, s, c string) {
	if !l.stdout {
		return
	}

	// Check verbosity
	log := false
	if vlevel, ok := verboseLevels[e]; ok {
		log = vlevel >= l.verbosity
	} else {
		log = l.verbosity > 3
	}

	if !log {
		return
	}

	now := time.Now().Format(l.tlayout)
	fmt.Printf("%s%s: %s%s: %s%s\n", Grey, now, c, strings.ToUpper(e), Reset, s)
	return
}

// Write log text specific path with filename [l.name]-[e].log and path l.path
func (l *Logger) writeToFile(e, s string) error {
	if !l.file {
		return errors.New("Write to file is disabled for this logger")
	}
	if err := checkPath(l.path); err != nil {
		return err
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
	l.mut.Lock()
	defer l.mut.Unlock()
	saveFile, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	defer saveFile.Close()

	_, err = saveFile.WriteString(errorStr)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
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
