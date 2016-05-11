Verbose - Logging for Go
=========================

[![GoDoc](https://godoc.org/github.com/dragonrider23/verbose?status.svg)](https://godoc.org/github.com/dragonrider23/verbose)

Verbose allows for organized, simplified, custom logging for any application. Verbose makes creating logs easy.
Create multiple loggers for different purposes each with their own handlers. Verbose supports both traditional
message only logs as well as structured logs. The handling of which is up to the implemented Handlers.

Usage
-----

Creating a Logger:

```Go
import "github.com/dragonrider23/verbose"

appLogger := verbose.New("app")
appLogger.Info("Application started")
appLogger.Warning("User not found")
appLogger.Error("I can't handle this")
appLogger.Fatal("Unhandled error occured") // Calls os.Exit(1)
```

You can also use the Get() func to get a specific logger. If Get() can't
find the logger by name, it will create a new logger:

```Go
logger := verbose.New("app")
logger = verbose.Get("app").Info("Error message") // Uses existing logger
logger = verbose.Get("module").Error("Error") // Creates new logger named 'module' and issues error
```

Supported Log Levels
--------------------

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

All functions take the form of Print[f]ln. E.g.: Print, Printf, Println.

Structured Logging
------------------

```go
logger.WithFields(verbose.Fields{
    "field 1": "value 1",
    "field 2": 42,
}).Debug("This is a debug message")
```

The fields should be formatted appropriately by the handler.

A Logger initially is nothing more than a shell. Without handlers it won't do anything.
Verbose comes with two prebuilt handlers. You can use your own handlers so long as they
satisfy the verbose.Handler interface. You can add a handler by calling `logger.AddHandler(name, Handler)`.
A Logger will cycle through all the handlers and send the message to any that report
they can handle the log level. Each handler should be given a unique name which can be used to later
remove to get the handler to make changes to it.

Included handlers
-----------------

StdoutHandler
-------------

The StdoutHandler will print colored log messages to stdout. This handler supports specifying
a minimum and maximum log level.

```go
sh := verbose.NewStdoutHandler()
```

FileHandler
-----------

The FileHandler will write log messages to a file or directory. If it's writing to a directory,
each log level will have its own file. Otherwise all log levels are written to a single file.
Like the StdoutHandler, FileHandlers support specifying a minimum and maximum log level to handle.

```go
// If path exists and is a file, it will write all logs to that file
// If path exists and is a directory, it will write logs to individual files per level
// If path does not exist but has an extension, assumed to be a file and attempts to create it
// If path does not exist and has no extension, assumed to be a directory and attempts to os.MkDirAll()
fh := verbose.NewFileHandler(path)
```

Release Notes
-------------

v2.0.0

- Added support for structured logging
- Removed LogLevelCustom
- Added [x]ln() functions to be compatible with the std lib logger

v1.0.0

- Initial Release

Versioning
----------

For transparency into the release cycle and in striving to maintain backward compatibility,
This application is maintained under the Semantic Versioning guidelines.
Sometimes I screw up, but I'll adhere to these rules whenever possible.

Releases will be numbered with the following format:

`<major>.<minor>.<patch>`

And constructed with the following guidelines:

- Breaking backward compatibility **bumps the major** while resetting minor and patch
- New additions without breaking backward compatibility **bumps the minor** while resetting the patch
- Bug fixes and misc changes **bumps only the patch**

For more information on SemVer, please visit <http://semver.org/>.

License
-------
This package is released under the terms of the MIT license. Please see LICENSE for more information.
