// This file was generated with level_generator. DO NOT EDIT

package verbose

import (
	"os"
	"fmt"
)

// Debug - Log Debug message
func (e *Entry) Debug(v ...interface{}) {
    e.log(LogLevelDebug, fmt.Sprint(v...))
    return
}

// Info - Log Info message
func (e *Entry) Info(v ...interface{}) {
    e.log(LogLevelInfo, fmt.Sprint(v...))
    return
}

// Notice - Log Notice message
func (e *Entry) Notice(v ...interface{}) {
    e.log(LogLevelNotice, fmt.Sprint(v...))
    return
}

// Warning - Log Warning message
func (e *Entry) Warning(v ...interface{}) {
    e.log(LogLevelWarning, fmt.Sprint(v...))
    return
}

// Error - Log Error message
func (e *Entry) Error(v ...interface{}) {
    e.log(LogLevelError, fmt.Sprint(v...))
    return
}

// Critical - Log Critical message
func (e *Entry) Critical(v ...interface{}) {
    e.log(LogLevelCritical, fmt.Sprint(v...))
    return
}

// Alert - Log Alert message
func (e *Entry) Alert(v ...interface{}) {
    e.log(LogLevelAlert, fmt.Sprint(v...))
    return
}

// Emergency - Log Emergency message
func (e *Entry) Emergency(v ...interface{}) {
    e.log(LogLevelEmergency, fmt.Sprint(v...))
    return
}

// Fatal - Log Fatal message
func (e *Entry) Fatal(v ...interface{}) {
    e.log(LogLevelFatal, fmt.Sprint(v...))
	os.Exit(1)
    return
}

// Panic - Log Panic message
func (e *Entry) Panic(v ...interface{}) {
    e.log(LogLevelEmergency, fmt.Sprint(v...))
    return
}

// Print - Log Print message
func (e *Entry) Print(v ...interface{}) {
    e.log(LogLevelInfo, fmt.Sprint(v...))
    return
}

// Printf friendly functions

// Debugf - Log formatted Debug message
func (e *Entry) Debugf(m string, v ...interface{}) {
    e.log(LogLevelDebug, fmt.Sprintf(m, v...))
    return
}

// Infof - Log formatted Info message
func (e *Entry) Infof(m string, v ...interface{}) {
    e.log(LogLevelInfo, fmt.Sprintf(m, v...))
    return
}

// Noticef - Log formatted Notice message
func (e *Entry) Noticef(m string, v ...interface{}) {
    e.log(LogLevelNotice, fmt.Sprintf(m, v...))
    return
}

// Warningf - Log formatted Warning message
func (e *Entry) Warningf(m string, v ...interface{}) {
    e.log(LogLevelWarning, fmt.Sprintf(m, v...))
    return
}

// Errorf - Log formatted Error message
func (e *Entry) Errorf(m string, v ...interface{}) {
    e.log(LogLevelError, fmt.Sprintf(m, v...))
    return
}

// Criticalf - Log formatted Critical message
func (e *Entry) Criticalf(m string, v ...interface{}) {
    e.log(LogLevelCritical, fmt.Sprintf(m, v...))
    return
}

// Alertf - Log formatted Alert message
func (e *Entry) Alertf(m string, v ...interface{}) {
    e.log(LogLevelAlert, fmt.Sprintf(m, v...))
    return
}

// Emergencyf - Log formatted Emergency message
func (e *Entry) Emergencyf(m string, v ...interface{}) {
    e.log(LogLevelEmergency, fmt.Sprintf(m, v...))
    return
}

// Fatalf - Log formatted Fatal message
func (e *Entry) Fatalf(m string, v ...interface{}) {
    e.log(LogLevelFatal, fmt.Sprintf(m, v...))
	os.Exit(1)
    return
}

// Panicf - Log formatted Panic message
func (e *Entry) Panicf(m string, v ...interface{}) {
    e.log(LogLevelEmergency, fmt.Sprintf(m, v...))
    return
}

// Printf - Log formatted Print message
func (e *Entry) Printf(m string, v ...interface{}) {
    e.log(LogLevelInfo, fmt.Sprintf(m, v...))
    return
}

// Println friendly functions

// Debugln - Log Debug message with newline
func (e *Entry) Debugln(v ...interface{}) {
    e.log(LogLevelDebug, e.sprintlnn(v...))
    return
}

// Infoln - Log Info message with newline
func (e *Entry) Infoln(v ...interface{}) {
    e.log(LogLevelInfo, e.sprintlnn(v...))
    return
}

// Noticeln - Log Notice message with newline
func (e *Entry) Noticeln(v ...interface{}) {
    e.log(LogLevelNotice, e.sprintlnn(v...))
    return
}

// Warningln - Log Warning message with newline
func (e *Entry) Warningln(v ...interface{}) {
    e.log(LogLevelWarning, e.sprintlnn(v...))
    return
}

// Errorln - Log Error message with newline
func (e *Entry) Errorln(v ...interface{}) {
    e.log(LogLevelError, e.sprintlnn(v...))
    return
}

// Criticalln - Log Critical message with newline
func (e *Entry) Criticalln(v ...interface{}) {
    e.log(LogLevelCritical, e.sprintlnn(v...))
    return
}

// Alertln - Log Alert message with newline
func (e *Entry) Alertln(v ...interface{}) {
    e.log(LogLevelAlert, e.sprintlnn(v...))
    return
}

// Emergencyln - Log Emergency message with newline
func (e *Entry) Emergencyln(v ...interface{}) {
    e.log(LogLevelEmergency, e.sprintlnn(v...))
    return
}

// Fatalln - Log Fatal message with newline
func (e *Entry) Fatalln(v ...interface{}) {
    e.log(LogLevelFatal, e.sprintlnn(v...))
	os.Exit(1)
    return
}

// Panicln - Log Panic message with newline
func (e *Entry) Panicln(v ...interface{}) {
    e.log(LogLevelEmergency, e.sprintlnn(v...))
    return
}

// Println - Log Print message with newline
func (e *Entry) Println(v ...interface{}) {
    e.log(LogLevelInfo, e.sprintlnn(v...))
    return
}
