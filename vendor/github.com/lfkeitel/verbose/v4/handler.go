package verbose

import "log"

// A Handler is an object that can be used by the Logger to log a message
type Handler interface {
	// Handles returns if it wants to handle a particular log level
	// This can be used to suppress the higher log levels in production.
	Handles(LogLevel) bool

	// WriteLog actually logs the message using any system the Handler wishes.
	// The Handler only needs to accept an Event.
	WriteLog(*Entry)

	// Close is used to give a handler a chance to close any open resources.
	Close()

	// SetFormatter replaces the handler's formatter with a specific one.
	SetFormatter(Formatter)

	// SetLevel will configure the handler to only handle a specific log level
	SetLevel(LogLevel)

	// SetMinLevel sets the minimum log level the handler will handle
	SetMinLevel(LogLevel)

	// SetMaxLevel sets the maximum log level the handler will handle
	SetMaxLevel(LogLevel)
}

// Won't compile if StdLogger can't be realized by a log.Logger
var (
	_ StdLogger = &log.Logger{}
	_ StdLogger = &Entry{}
	_ StdLogger = &Logger{}
)

// StdLogger is what your verbose-enabled library should take, that way
// it'll accept a stdlib logger and a verbose logger. There's no standard
// interface, this is the closest we get, unfortunately.
type StdLogger interface {
	Print(...interface{})
	Printf(string, ...interface{})
	Println(...interface{})

	Fatal(...interface{})
	Fatalf(string, ...interface{})
	Fatalln(...interface{})

	Panic(...interface{})
	Panicf(string, ...interface{})
	Panicln(...interface{})
}
