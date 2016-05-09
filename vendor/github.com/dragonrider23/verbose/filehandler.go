package verbose

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

// FileHandler writes log messages to a file to a directory
type FileHandler struct {
	min      LogLevel
	max      LogLevel
	path     string
	separate bool
	m        sync.Mutex
}

// NewFileHandler takes the path and returns a FileHandler. If the path exists,
// file or directory mode will be Determined by what path is. If it doesn't exist,
// the mode will be file if path has an extension, otherwise it will be directory.
// In file mode, all log messages are written to a single file.
// In directory mode, each level is written to it's own file.
func NewFileHandler(path string) (*FileHandler, error) {
	f := &FileHandler{
		min:  LogLevelDebug,
		max:  LogLevelEmergency,
		path: path,
		m:    sync.Mutex{},
	}

	// Determine of the path is a file or directory
	// We cannot assume the path exists yet
	stat, err := os.Stat(path)
	if err == nil { // Easiest, path exists
		f.separate = stat.IsDir()
	} else if os.IsNotExist(err) {
		// Typically an extension means it's a file
		ext := filepath.Ext(path)
		if ext == "" {
			// Attempt to create the directory
			if err := os.MkdirAll(path, 0755); err != nil {
				return nil, err
			}
			f.separate = true
		} else {
			// Attempt to create the directory
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return nil, err
			}
			// Attempt to create the file
			file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
			if err != nil {
				return nil, err
			}
			file.Close()
			f.separate = false
		}
	}
	return f, nil
}

// SetLevel will set both the minimum and maximum log levels to l. This makes
// the handler only respond to the single level l.
func (f *FileHandler) SetLevel(l LogLevel) {
	f.min = l
	f.max = l
}

// SetMinLevel will set the minimum log level the handler will handle.
func (f *FileHandler) SetMinLevel(l LogLevel) {
	if l > f.max {
		return
	}
	f.min = l
}

// SetMaxLevel will set the maximum log level the handler will handle.
func (f *FileHandler) SetMaxLevel(l LogLevel) {
	if l < f.min {
		return
	}
	f.max = l
}

// Handles returns whether the handler handles log level l.
func (f *FileHandler) Handles(l LogLevel) bool {
	return (f.min <= l && l <= f.max)
}

// WriteLog will write the log message to a file.
func (f *FileHandler) WriteLog(e *Entry) {
	var logfile string
	if !f.separate {
		logfile = f.path
	} else {
		logfile = fmt.Sprintf("%s-%s.log", strings.ToLower(e.Level.String()), e.Logger.Name())
		logfile = path.Join(f.path, logfile)
	}

	buf := &bytes.Buffer{}
	fmt.Fprintf(
		buf,
		"%s: %s: %s: %s",
		e.Timestamp.Format("2006-01-02 15:04:05 MST"),
		strings.ToUpper(e.Level.String()),
		e.Logger.Name(),
		e.Message,
	)
	for k, v := range e.Data {
		fmt.Fprintf(buf, " %s=%v", k, v)
	}
	buf.WriteByte('\n')

	f.m.Lock()
	defer f.m.Unlock()

	file, err := os.OpenFile(logfile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
	}
	defer file.Close()

	_, err = file.Write(buf.Bytes())
	if err != nil {
		fmt.Printf("Error writing to log file: %v\n", err)
	}
	return
}

// Close satisfies the interface, NOOP
func (f *FileHandler) Close() {}
