# Verbose - Logging for Go

[![GoDoc](https://godoc.org/github.com/lfkeitel/verbose?status.svg)](https://godoc.org/github.com/lfkeitel/verbose)

Verbose allows for organized, simplified, custom logging for any application. Verbose makes creating logs easy.
Create multiple loggers for different purposes each with their own handlers. Verbose supports both traditional
message only logs as well as structured logs. The handling of which is up to the implemented Handlers.

## Usage

Creating a Logger:

```Go
import "github.com/lfkeitel/verbose"

appLogger := verbose.New("app")
appLogger.Info("Application started")
appLogger.Warning("User not found")
appLogger.Error("I can't handle this")
appLogger.Fatal("Unhandled error occurred") // Calls os.Exit(1)
```

You can also use the Get() func to get a specific logger. If Get() can't
find the logger by name, it will create a new logger:

```Go
logger := verbose.New("app")
logger = verbose.Get("app").Info("Error message") // Uses existing logger
logger = verbose.Get("module").Error("Error") // Creates new logger named 'module' and issues error
```

## Supported Log Levels

- Debug
- Info
- Notice
- Warning
- Error
- Critical
- Alert
- Emergency
- Fatal (calls os.Exit(1))

You can also use the following functions:

- Print
- Panic

All functions take the form of Print[f|ln]. E.g.: Print, Printf, Println.

## Structured Logging

```go
logger.WithField("field 1", data)

logger.WithFields(verbose.Fields{
    "field 1": "value 1",
    "field 2": 42,
}).Debug("This is a debug message")
```

The fields should be formatted appropriately by the handler.

## Handlers

A Logger initially is nothing more than a shell. Without handlers it won't do anything.
Verbose comes with two pre-built handlers. You can use your own handlers so long as they
satisfy the verbose.Handler interface. You can add a handler by calling `logger.AddHandler(name, Handler)`.
A Logger will cycle through all the handlers and send the message to any that report
they can handle the log level. Each handler should be given a unique name which can be used to later
remove or get the handler to make changes to it.

### StdoutHandler

The StdoutHandler will print colored log messages to stdout.

```go
// Handler that supports color terminals
sh := verbose.NewStdoutHandler(true)

// Handler that doesn't use color
sh := verbose.NewStdoutHandler(false)
```

### FileHandler

The FileHandler will write log messages to a file or directory. If it's writing to a directory,
each log level will have its own file. Otherwise all log levels are written to a single file.

```go
// If path exists and is a file, it will write all logs to that file
// If path exists and is a directory, it will write logs to individual files per level
// If path does not exist but has an extension, assumed to be a file and attempts to create it
// If path does not exist and has no extension, assumed to be a directory and attempts to os.MkDirAll()
fh := verbose.NewFileHandler(path)
```

## Formatters

A formatter is used to actually construct a log line that a handler will then store or display.
Like handlers, Verbose comes with 3 pre-built formatters but anything satisfying the interface
will work.

Each handler has a default formatter. The File and StdOut handlers use the LineFormatter as
their defaults. To change a formatter, use the Handler.SetFormatter(Formatter) method.

### Time Format

The time format used by formatters can be set using the Formatter.SetTimeFormat() method.
The default time format for included formatters is RFC3339: "2006-01-02T15:04:05Z07:00". The time
format can be any valid Go time format.

### JSONFormatter

The JSON formatter is great when the logs are being processed by a centralized logging solution
or some other computerized system. It will generate a JSON object with the following structure:

```json
{
    "timestamp": "1970-01-01T12:00:00Z",
    "level": "INFO",
    "logger": "app",
    "message": "Hello, world",
    "data": {
        "field 1": "data 1",
        "field 2": "data 2"
    }
}
```

Any structured fields will go in the data object.

### LineFormatter

The line formatter is designed to be human readable either for file that will mainly be viewed by
humans, or for standard output. A sample output line would be:

```
1970-01-01T12:00:00Z: INFO: app: message: | "field 1": "value 1", "field 2": "value 2"
```

### ColoredLineFormatter

Same as the line formatter but uses ASCII color codes to make things pretty. This formatter is really
only meant for standard output as the escape codes are really annoying when looking at a log file.

## Release Notes

v4.0.0

- Expanded Formatter interface
    - SetTimeFormat(string)
- Added generator functions for Formatters
    - NewJSONFormatter()
    - NewLineFormatter()
    - NewColoredLineFormatter()
- Use RFC3339 as the default time format

v3.0.0

- Expanded Handler interface
    - SetFormatter(Formatter)
    - SetLevel(LogLevel)
    - SetMinLevel(LogLevel)
    - SetMaxLevel(LogLevel)
- Added support for formatters
    - Included formatters:
        - JSON
        - Line
        - Line with Color
- Use Fatal as the default Handler max for StdOut and FileHandlers

v2.0.0

- Added support for structured logging
- Removed LogLevelCustom
- Added [x]ln() functions to be compatible with the std lib logger

v1.0.0

- Initial Release

## Versioning

For transparency into the release cycle and in striving to maintain backward compatibility,
this application is maintained under the Semantic Versioning guidelines.
Sometimes I screw up, but I'll adhere to these rules whenever possible.

Releases will be numbered with the following format:

`<major>.<minor>.<patch>`

And constructed with the following guidelines:

- Breaking backward compatibility **bumps the major** while resetting minor and patch
- New additions without breaking backward compatibility **bumps the minor** while resetting the patch
- Bug fixes and misc changes **bumps only the patch**

For more information on SemVer, please visit <http://semver.org/>.

## License

This package is released under the terms of the MIT license. Please see LICENSE for more information.
