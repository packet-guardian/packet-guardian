package verbose

import (
	"bytes"
	"fmt"
	"strings"
)

type Formatter interface {
	Format(*Entry) string
	FormatByte(*Entry) []byte
}

type JSONFormatter struct{}

func (j *JSONFormatter) Format(e *Entry) string {
	return string(j.FormatByte(e))
}

func (j *JSONFormatter) FormatByte(e *Entry) []byte {
	buf := bytes.Buffer{}
	buf.WriteByte('{')
	buf.WriteString(fmt.Sprintf(`"timestamp":"%s",`, e.Timestamp.Format("2006-01-02 15:04:05 MST")))
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

type LineFormatter struct{}

func (l *LineFormatter) Format(e *Entry) string {
	return string(l.FormatByte(e))
}

func (l *LineFormatter) FormatByte(e *Entry) []byte {
	buf := &bytes.Buffer{}
	fmt.Fprintf(
		buf,
		"%s: %s: %s: %s",
		e.Timestamp.Format("2006-01-02 15:04:05 MST"),
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

type ColoredLineFormatter struct{}

func (l *ColoredLineFormatter) Format(e *Entry) string {
	return string(l.FormatByte(e))
}

func (l *ColoredLineFormatter) FormatByte(e *Entry) []byte {
	buf := &bytes.Buffer{}
	fmt.Fprintf(
		buf,
		"%s%s: %s%s: %s%s: %s%s",
		ColorGrey,
		e.Timestamp.Format("2006-01-02 15:04:05 MST"),
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
