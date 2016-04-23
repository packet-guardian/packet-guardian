Verbose - Logging for Go
=========================

[![GoDoc](https://godoc.org/github.com/dragonrider23/verbose?status.svg)](https://godoc.org/github.com/dragonrider23/verbose)

Verbose allows for organized, simplified, custom
logging for any application. Verbose makes creating logs easy.
Create multiple loggers for different purposes each with their
own handlers.

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

To specify your own error level, use the Log function:

```Go
appLogger.Log(verbose.LogLevelCustom, "This is a custom log level message")
```

You can also use the Get() func to get a specific logger. If Get() can't
find the logger by name, it will create a new logger:

```Go
logger := verbose.New("app")
logger = verbose.Get("app").Info("Error message") // Uses existing logger
logger = verbose.Get("module").Error("Error") // Creates new logger named 'module' and issues error
```

A Logger initially is nothing more than a shell. Without handlers it won't do anything.
Verbose comes with two prebuilt handlers. You can use your own handlers so long as they
satisfy the verbose.Handler interface. You can add a handler by calling `logger.AddHandler(Handler)`.
A Logger will cycle through all the handlers and send the message to any that report
they can handle the log level.

Included handlers
-----------------

StdoutHandler
-------------

The StdoutHandler will print colored log messages to stdout. This handler supports specifying
a minimum and maximum log level.

FileHandler
-----------

The FileHandler will write log messages to a file or directory. If it's writing to a directory,
each log level will have its own file. Otherwise all log levels are written to a single file.
Like the StdoutHandler, FileHandlers support specifying a minimum and maximum log level to handle.

Release Notes
-------------

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
