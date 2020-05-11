# logger

package logger contains LoggerWrapper, BasicLogger and logging interfaces.


### LoggerWrapper
In order not to extort any specific logging package, a logger wraper has been created.
LoggerWrapper wraps around third-party loggers that implement one of the logging-interfaces:
```
	- StdLogger - standard library logger interface
	- LeveledLogger - basic leveled logger interface
	- DebugLeveledLogger - LeveledLogger with the debug2 and debug3 support
	- ShortLeveledLogger - basic leveled logger interfaces with shortened method names
	- ExtendedLeveledLogger - a fully leveled logger interface
```
This solution allows to use ExtendedLeveledLogger interface methods for most of the third-party
logging packages.

#### Wrapping third-party logger
```go
import (
	"github.com/neuronlabs/uni-logger"
	"some/loggingpkg"
)

func main(){
	// Having a third-party logger that implements any of the package interfaces
	var myLogger *logginpkg.Logger = loggingpkg.New()

	// If 'myLogger' doesn't implement 'ExtendedLeveledLogger' but any generic function
	// uses logs of that interface, it can be wrapped using LoggerWrapper.

	// if 'myLogger' doesn't implement any of the listed interfaces MustGetLoggingWrapper would panic.
	var wrappedLoggerMust *unilogger.LoggerWrapper = unilogger.MustGetLoggerWrapped(myLogger)

	// The other function to get a logging wrapper is NewLoggingWrapper(myLogger) which returns 
	// new *LoggerWrapper or an error if it doesn't implement listed interfaces.
	var wrappedLoggerNew *unilogger.LoggerWrapper
	var err error
	wrappedLoggerNew, err = unilogger.NewLoggingWrapper(myLogger)
	if err != nil {
		...
	}

	wrappedLoggerNew.Infoln("It works!")
}
```


### BasicLogger
The package contains also BasicLogger that implements 'LeveledLogger' interface.
It is very simple and lightweight implementation of leveled logger.
```go
import (
	"log"
	"os"
	"github.com/neuronlabs/uni-logger"
)
		


func main(){
	// BasicLogger is simple leveled logger that implements LeveledLogger interface.
	var basicLogger *logger.BasicLogger

	// In order to get new basic logger use NewBasicLogger() function
	basicLogger = logger.NewBasicLogger(os.Stderr, "", log.Ltime)

	// BasicLogger implements LeveledLogger interface
	basicLogger.Error("This should log an error.")
}
```

### Log Levels
The package uses 8 basic log levels. 
```go
// Level defines a logging level used in BasicLogger
type Level int

// Following levels are supported in BasicLogger
const (
	DEBUG3 Level = iota
	DEBUG2
	DEBUG 
	INFO
	WARNING
	ERROR
	CRITICAL
	PRINT
)
```
The BasicLogger allows to set a logging level so that no lower level logs would be printed.
This allows to control the logging output just for specified level (or higher).

