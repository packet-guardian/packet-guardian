// This file was generated with level_generator. DO NOT EDIT

package verbose

import "os"

// LogLevel is used to compare levels in a consistant manner
type LogLevel int

// String returns the stringified version of LogLevel.
// I.e., "Error" for LogLevelError, and "Debug" for LogLevelDebug
// It will return an empty string for any undefined level.
func (l LogLevel) String() string {
	if s, ok := levelString[l]; ok {
		return s
	}
	return ""
}

// These are the defined log levels
const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelNotice
	LogLevelWarning
	LogLevelError
	LogLevelCritical
	LogLevelAlert
	LogLevelEmergency
	LogLevelFatal
)

// LogLevel to stringified versions
var levelString = map[LogLevel]string{ 
	LogLevelDebug:     "Debug",
	LogLevelInfo:     "Info",
	LogLevelNotice:     "Notice",
	LogLevelWarning:     "Warning",
	LogLevelError:     "Error",
	LogLevelCritical:     "Critical",
	LogLevelAlert:     "Alert",
	LogLevelEmergency:     "Emergency",
	LogLevelFatal:     "Fatal",
}

// Debug - Log Debug message
func (l *Logger) Debug(v ...interface{}) {
    NewEntry(l).Debug(v...)
    return
}

// Info - Log Info message
func (l *Logger) Info(v ...interface{}) {
    NewEntry(l).Info(v...)
    return
}

// Notice - Log Notice message
func (l *Logger) Notice(v ...interface{}) {
    NewEntry(l).Notice(v...)
    return
}

// Warning - Log Warning message
func (l *Logger) Warning(v ...interface{}) {
    NewEntry(l).Warning(v...)
    return
}

// Error - Log Error message
func (l *Logger) Error(v ...interface{}) {
    NewEntry(l).Error(v...)
    return
}

// Critical - Log Critical message
func (l *Logger) Critical(v ...interface{}) {
    NewEntry(l).Critical(v...)
    return
}

// Alert - Log Alert message
func (l *Logger) Alert(v ...interface{}) {
    NewEntry(l).Alert(v...)
    return
}

// Emergency - Log Emergency message
func (l *Logger) Emergency(v ...interface{}) {
    NewEntry(l).Emergency(v...)
    return
}

// Fatal - Log Fatal message
func (l *Logger) Fatal(v ...interface{}) {
    NewEntry(l).Fatal(v...)
	os.Exit(1)
    return
}

// Panic - Log Panic message
func (l *Logger) Panic(v ...interface{}) {
    NewEntry(l).Panic(v...)
    return
}

// Print - Log Print message
func (l *Logger) Print(v ...interface{}) {
    NewEntry(l).Print(v...)
    return
}

// Printf friendly functions

// Debugf - Log formatted Debug message
func (l *Logger) Debugf(m string, v ...interface{}) {
	NewEntry(l).Debugf(m, v...)
    return
}

// Infof - Log formatted Info message
func (l *Logger) Infof(m string, v ...interface{}) {
	NewEntry(l).Infof(m, v...)
    return
}

// Noticef - Log formatted Notice message
func (l *Logger) Noticef(m string, v ...interface{}) {
	NewEntry(l).Noticef(m, v...)
    return
}

// Warningf - Log formatted Warning message
func (l *Logger) Warningf(m string, v ...interface{}) {
	NewEntry(l).Warningf(m, v...)
    return
}

// Errorf - Log formatted Error message
func (l *Logger) Errorf(m string, v ...interface{}) {
	NewEntry(l).Errorf(m, v...)
    return
}

// Criticalf - Log formatted Critical message
func (l *Logger) Criticalf(m string, v ...interface{}) {
	NewEntry(l).Criticalf(m, v...)
    return
}

// Alertf - Log formatted Alert message
func (l *Logger) Alertf(m string, v ...interface{}) {
	NewEntry(l).Alertf(m, v...)
    return
}

// Emergencyf - Log formatted Emergency message
func (l *Logger) Emergencyf(m string, v ...interface{}) {
	NewEntry(l).Emergencyf(m, v...)
    return
}

// Fatalf - Log formatted Fatal message
func (l *Logger) Fatalf(m string, v ...interface{}) {
	NewEntry(l).Fatalf(m, v...)
	os.Exit(1)
    return
}

// Panicf - Log formatted Panic message
func (l *Logger) Panicf(m string, v ...interface{}) {
	NewEntry(l).Panicf(m, v...)
    return
}

// Printf - Log formatted Print message
func (l *Logger) Printf(m string, v ...interface{}) {
	NewEntry(l).Printf(m, v...)
    return
}

// Println friendly functions

// Debugln - Log Debug message with newline
func (l *Logger) Debugln(v ...interface{}) {
    NewEntry(l).Debugln(v...)
    return
}

// Infoln - Log Info message with newline
func (l *Logger) Infoln(v ...interface{}) {
    NewEntry(l).Infoln(v...)
    return
}

// Noticeln - Log Notice message with newline
func (l *Logger) Noticeln(v ...interface{}) {
    NewEntry(l).Noticeln(v...)
    return
}

// Warningln - Log Warning message with newline
func (l *Logger) Warningln(v ...interface{}) {
    NewEntry(l).Warningln(v...)
    return
}

// Errorln - Log Error message with newline
func (l *Logger) Errorln(v ...interface{}) {
    NewEntry(l).Errorln(v...)
    return
}

// Criticalln - Log Critical message with newline
func (l *Logger) Criticalln(v ...interface{}) {
    NewEntry(l).Criticalln(v...)
    return
}

// Alertln - Log Alert message with newline
func (l *Logger) Alertln(v ...interface{}) {
    NewEntry(l).Alertln(v...)
    return
}

// Emergencyln - Log Emergency message with newline
func (l *Logger) Emergencyln(v ...interface{}) {
    NewEntry(l).Emergencyln(v...)
    return
}

// Fatalln - Log Fatal message with newline
func (l *Logger) Fatalln(v ...interface{}) {
    NewEntry(l).Fatalln(v...)
	os.Exit(1)
    return
}

// Panicln - Log Panic message with newline
func (l *Logger) Panicln(v ...interface{}) {
    NewEntry(l).Panicln(v...)
    return
}

// Println - Log Print message with newline
func (l *Logger) Println(v ...interface{}) {
    NewEntry(l).Println(v...)
    return
}
