package verbose

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

// Color is an escaped color code for the terminal
type Color string

// Pre-defined colors
const (
	ColorReset   Color = "\033[0m"
	ColorRed     Color = "\033[31m"
	ColorGreen   Color = "\033[32m"
	ColorYellow  Color = "\033[33m"
	ColorBlue    Color = "\033[34m"
	ColorMagenta Color = "\033[35m"
	ColorCyan    Color = "\033[36m"
	ColorWhite   Color = "\033[37m"
	ColorGrey    Color = "\033[90m"
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
		max: LogLevelEmergency,
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
func (s *StdoutHandler) WriteLog(e *Entry) {
	buf := &bytes.Buffer{}
	now := e.Timestamp.Format("2006-01-02 15:04:05 MST")
	fmt.Fprintf(
		buf,
		"%s%s: %s%s: %s%s: %s%s",
		ColorGrey,
		now,
		colors[e.Level],
		strings.ToUpper(e.Level.String()),
		ColorGreen,
		e.Logger.Name(),
		ColorReset,
		e.Message,
	)
	for k, v := range e.Data {
		fmt.Fprintf(buf, " %s=%v", k, v)
	}
	buf.WriteByte('\n')
	fmt.Fprint(s.out, buf.String())
}

// Close satisfies the interface, NOOP
func (s *StdoutHandler) Close() {}
