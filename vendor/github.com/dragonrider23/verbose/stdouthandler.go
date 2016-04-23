package verbose

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

var colors = map[LogLevel]Color{
	LogLevelDebug:     ColorBlue,
	LogLevelInfo:      ColorCyan,
	LogLevelNotice:    ColorCyan,
	LogLevelWarning:   ColorMagenta,
	LogLevelError:     ColorRed,
	LogLevelCritical:  ColorRed,
	LogLevelAlert:     ColorRed,
	LogLevelEmergency: ColorRed,
	LogLevelFatal:     ColorRed,
	LogLevelCustom:    ColorWhite,
}

// StdoutHandler writes log message to standard out
// It even uses color!
type StdoutHandler struct {
	min LogLevel
	max LogLevel
	out io.Writer // Usually os.Stdout, mainly used for testing
}

// NewStdoutHandler creates a new StdoutHandler, surprise!
func NewStdoutHandler() *StdoutHandler {
	return &StdoutHandler{
		min: LogLevelDebug,
		max: LogLevelCustom,
		out: os.Stdout,
	}
}

// SetLevel will set both the minimum and maximum log levels to l. This makes
// the handler only respond to the single level l.
func (s *StdoutHandler) SetLevel(l LogLevel) {
	s.min = l
	s.max = l
}

// SetMinLevel will set the minimum log level the handler will handle.
func (s *StdoutHandler) SetMinLevel(l LogLevel) {
	if l > s.max {
		return
	}
	s.min = l
}

// SetMaxLevel will set the maximum log level the handler will handle.
func (s *StdoutHandler) SetMaxLevel(l LogLevel) {
	if l < s.min {
		return
	}
	s.max = l
}

// Handles returns whether the handler handles log level l.
func (s *StdoutHandler) Handles(l LogLevel) bool {
	return (s.min <= l && l <= s.max)
}

// WriteLog writes the log message to standard output
func (s *StdoutHandler) WriteLog(l LogLevel, name, msg string) {
	now := time.Now().Format("2006-01-02 15:04:05 MST")
	fmt.Fprintf(
		s.out,
		"%s%s: %s%s: %s%s: %s%s\n",
		ColorGrey,
		now,
		colors[l],
		strings.ToUpper(l.String()),
		ColorGreen,
		name,
		ColorReset,
		msg,
	)
}

// Close satisfies the interface, NOOP
func (s *StdoutHandler) Close() {}
