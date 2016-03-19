Logging Package for Go
======================

Go-logger allows for organized, simplified, custom
logging for any application. Go-logger makes creating logs easy.
Create multiple loggers for different purposes each with their
own specific settings such as verbosity, log file location, and
whether it's show on stdout, written to file, or only one of them.
Go-logger comes with 4 pre-made logging levels ready for use. There's
also a generic Log() function that allows for custom log types.

Go-logger also has timers that can be attached to logs. This makes it
easy to track how long a function or request takes to run. Go-logger
also comes with a wrapper function to check for errors in an application.
Instead of having to write out the log to file code for every err if
statement, CheckError() will take the error, check if one exists and
then write it to the Go-logger of choice. It will also return boolean
to tell the calling function if an error happened or not so any custom
actions can be performed if needed.

Documentation
-------------

[View Package Documentation](http://godoc.org/github.com/dragonrider23/go-logger)

Usage
-----

Logs are by default written to folder "./logs" relative to where the binary was executed.

Basic usage:

```Go
import logger "github.com/dragonrider23/go-logger"

appLogger := logger.New("app")
appLogger.Info("Application started")
appLogger.Warning("User not found")
appLogger.Error("I can't handle this")
appLogger.Fatal("Unhandled error occured") // Calls os.Exit(1)
```

To specify your own error level, use the Log function:

```Go
appLogger.Log("Notice", blue, "%s", "This is a notice", ...interface{})
```

You can also use the Get() func to call a specific logger. If Get() can't
find the logger by name, it will create a new logger:

```Go
logger.New("app")
logger.Get("app").Info("Error message") // Uses existing logger
logger.Get("module").Error("Error") // Creates new logger named 'module' and issues error
```

Each logger is unique by name so you can specify which loggers can print to stdout,
where each logger saves its log files, if the logger even writes to a file,
 as well as the timestampe format and verbosity:

```Go
appLogger.Path("path/to/log/files") // Changes log file path
appLogger.NoStdout() // Disables stdout printing for the logger
appLogger.NoFile() // Disables logger from writting to a file
appLogger.Stdout() // Enables stdout printing for the logger (enabled by default)
appLogger.File() // Enables logger writting to a file (enabled by default)
appLogger.TimeCode("2006-01-02 15:04:05 MST") // Changes timestamp format
appLogger.Verbose(2) // Sets verbose level to 2 (Error and Fatal)
```

You can even chain multiple configuration functions together:

```Go
logger.New("request").Path("reqlogs").NoStdout()
```

Lastly, if you want to get rid of the logger (using the Get() syntax):

```Go
logger.Get("app").Close()
```

Logger also supports timers. Each logger can only have one timer. StopTimer() will return
the elapsed time as a string:

```Go
logger.Get("app").StartTimer() // Starts timer for logger "app"
logger.Get("app").StopTimer("Action took {time} to complete") // Stops timer and logs message.
{time} is replaced by the elapsed time
elapsed := logger.Get("app").StopTimer("") // Doesn't write an Info log if string is empty
```

You can set the global verbosity level of stdout with Verbose(). This is used when a logger
is created to specify its verbosity level:

```Go
logger.Verbose(3) // All log levels inclusing custom levels
logger.Verbose(2) // Warning, Error, Fatal (Default)
logger.Verbose(1) // Error, Fatal
logger.Verbose(0) // Fatal
```

Available Colors for Logs
-------------------------

* Reset
* Red
* Green
* Yellow
* Blue
* Magenta
* Cyan
* White
* Grey

Release Notes
-------------

v2.2.0

- Added StdOut() and File() which will reenable StdOut and file writing for logger
- Added the Green and Yellow ANSI colors
- Code cleanup

v2.1.0

- Logger specific verbosity level
- Wrapper function for error checking
- The logger name is used in the filename
- Improved documentation

v2.0.0

- Timers on logger
- Error() is a log level
- Log() replaces Error() as custom log level
- New() will return an existing logger if the name is created
- Get() will create a logger if the name doesn't exist
- Bump major version because of Error() Log() change
- Set verbosity for stdout

v1.0.0

- Initial Release

Versioning
----------

For transparency into the release cycle and in striving to maintain backward compatibility,
This application is maintained under the Semantic Versioning guidelines.
Sometimes we screw up, but we'll adhere to these rules whenever possible.

Releases will be numbered with the following format:

`<major>.<minor>.<patch>`

And constructed with the following guidelines:

- Breaking backward compatibility **bumps the major** while resetting minor and patch
- New additions without breaking backward compatibility **bumps the minor** while resetting the patch
- Bug fixes and misc changes **bumps only the patch**

For more information on SemVer, please visit <http://semver.org/>.

License
-------
This package is released under the terms of the MIT license. Please see LICENSE.md for more information.
