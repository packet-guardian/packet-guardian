package verbose

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

type Formatter interface {
	Format(*Entry) string
	FormatByte(*Entry) []byte
	SetTimeFormat(string)
}

type JSONFormatter struct {
	timeFormat string
}

func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{
		timeFormat: time.RFC3339,
	}
}

func (j *JSONFormatter) Format(e *Entry) string {
	return string(j.FormatByte(e))
}

func (j *JSONFormatter) FormatByte(e *Entry) []byte {
	buf := bytes.Buffer{}
	buf.WriteByte('{')
	buf.WriteString(fmt.Sprintf(`"timestamp":"%s",`, e.Timestamp.Format(j.timeFormat)))
	buf.WriteString(fmt.Sprintf(`"level":"%s",`, strings.ToUpper(e.Level.String())))
	buf.WriteString(fmt.Sprintf(`"logger":"%s",`, e.Logger.Name()))
	buf.WriteString(fmt.Sprintf(`"message":"%s",`, e.Message))
	buf.WriteString(`"data":{`)
	dataLen := len(e.Data)
	for k, v := range e.Data {
		buf.WriteString(fmt.Sprintf(`"%s":"%v"`, k, v))
		if dataLen > 1 {
			buf.WriteByte(',')
		}
		dataLen--
	}
	buf.WriteByte('}') // End data key
	buf.WriteByte('}') // End complete object
	buf.WriteByte('\n')
	return buf.Bytes()
}

func (j *JSONFormatter) SetTimeFormat(f string) {
	j.timeFormat = f
}

type LineFormatter struct {
	timeFormat string
}

func NewLineFormatter() *LineFormatter {
	return &LineFormatter{
		timeFormat: time.RFC3339,
	}
}

func (l *LineFormatter) Format(e *Entry) string {
	return string(l.FormatByte(e))
}

func (l *LineFormatter) FormatByte(e *Entry) []byte {
	buf := &bytes.Buffer{}
	fmt.Fprintf(
		buf,
		"%s: %s: %s: %s",
		e.Timestamp.Format(l.timeFormat),
		strings.ToUpper(e.Level.String()),
		e.Logger.Name(),
		e.Message,
	)
	dataLen := len(e.Data)
	if dataLen > 0 {
		buf.WriteString(" |")
		for k, v := range e.Data {
			fmt.Fprintf(buf, ` "%s": "%v"`, k, v)
			if dataLen > 1 {
				buf.WriteByte(',')
			}
			dataLen--
		}
	}
	buf.WriteByte('\n')
	return buf.Bytes()
}

func (l *LineFormatter) SetTimeFormat(f string) {
	l.timeFormat = f
}

type ColoredLineFormatter struct {
	timeFormat string
}

func NewColoredLineFormatter() *ColoredLineFormatter {
	return &ColoredLineFormatter{
		timeFormat: time.RFC3339,
	}
}

func (l *ColoredLineFormatter) Format(e *Entry) string {
	return string(l.FormatByte(e))
}

func (l *ColoredLineFormatter) FormatByte(e *Entry) []byte {
	buf := &bytes.Buffer{}
	fmt.Fprintf(
		buf,
		"%s%s: %s%s: %s%s: %s%s",
		ColorGrey,
		e.Timestamp.Format(l.timeFormat),
		colors[e.Level],
		strings.ToUpper(e.Level.String()),
		ColorGreen,
		e.Logger.Name(),
		ColorReset,
		e.Message,
	)
	dataLen := len(e.Data)
	if dataLen > 0 {
		buf.WriteString(" |")
		for k, v := range e.Data {
			fmt.Fprintf(buf, ` "%s": "%v"`, k, v)
			if dataLen > 1 {
				buf.WriteByte(',')
			}
			dataLen--
		}
	}
	buf.WriteByte('\n')
	return buf.Bytes()
}

func (l *ColoredLineFormatter) SetTimeFormat(f string) {
	l.timeFormat = f
}
