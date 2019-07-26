package log

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/neuronlabs/uni-logger"

	"github.com/neuronlabs/errors"
	"github.com/neuronlabs/neuron-core/class"
)

const (
	// LDEBUG3 is the logger DEBUG3 level.
	LDEBUG3 = unilogger.DEBUG3
	// LDEBUG2 is the logger DEBUG2 level.
	LDEBUG2 = unilogger.DEBUG2
	// LDEBUG is the logger DEBUG level.
	LDEBUG = unilogger.DEBUG
	// LINFO is the logger INFO level.
	LINFO = unilogger.INFO
	// LWARNING is the logger WARNING level.
	LWARNING = unilogger.WARNING
	// LERROR is the logger ERROR level.
	LERROR = unilogger.ERROR
	// LCRITICAL is the logger CRITICAL level.
	LCRITICAL = unilogger.CRITICAL
	// LPRINT is the logger PRINT level.
	LPRINT = unilogger.PRINT
	// LUNKNOWN is the unspecified logger level.
	LUNKNOWN = unilogger.UNKNOWN
)

var (
	logger         unilogger.LeveledLogger
	currentLevel   = LINFO
	debugLeveled   unilogger.DebugLeveledLogger
	isDebugLeveled bool
)

// Default creates and sets new unilogger.BasicLogger with writer to 'os.Stderr'.
func Default() {
	basic := unilogger.NewBasicLogger(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile)
	basic.SetOutputDepth(4)
	SetLogger(basic)
}

// Debug writes the LDEBUG level log.
func Debug(args ...interface{}) {
	if logger != nil {
		logger.Debug(args...)
	}
}

// Debug2 writes the LDEBUG2 level logs.
func Debug2(args ...interface{}) {
	if !isDebugLeveled {
		if logger != nil {
			logger.Debug(args...)
		}
	} else {
		if debugLeveled != nil {
			debugLeveled.Debug2(args...)
		}
	}
}

// Debug2f writes the formated LDEBUG2 level log.
func Debug2f(format string, args ...interface{}) {
	if !isDebugLeveled {
		if logger != nil {
			logger.Debugf(format, args...)
		}
	} else {
		if debugLeveled != nil {
			debugLeveled.Debug2f(format, args...)
		}
	}
}

// Debug3 writes the LDEBUG3 level logs.
func Debug3(args ...interface{}) {
	if !isDebugLeveled {
		if logger != nil {
			logger.Debug(args...)
		}
	} else {
		if debugLeveled != nil {
			debugLeveled.Debug3(args...)
		}
	}
}

// Debug3f writes the formated LDEBUG2 level log.
func Debug3f(format string, args ...interface{}) {
	if !isDebugLeveled {
		if logger != nil {
			logger.Debugf(format, args...)
		}
	} else {
		if debugLeveled != nil {
			debugLeveled.Debug3f(format, args...)
		}
	}
}

// Debugf writes the formated LDEBUG level log.
func Debugf(format string, args ...interface{}) {
	if logger != nil {
		logger.Debugf(format, args...)
	}
}

// Error writes the LERROR level log.
func Error(args ...interface{}) {
	if logger != nil {
		logger.Error(args...)
	}
}

// Errorf writes the formated LERROR level log.
func Errorf(format string, args ...interface{}) {
	if logger != nil {
		logger.Errorf(format, args...)
	}
}

// Fatal writes the fatal - LCRITICAL level log.
func Fatal(args ...interface{}) {
	if logger != nil {
		logger.Fatal(args...)
	}
}

// Fatalf writes the formated fatal - LCRITICAL level log.
func Fatalf(format string, args ...interface{}) {
	if logger != nil {
		logger.Fatalf(format, args...)
	} else {
		fmt.Printf(format, args...)
		os.Exit(1)
	}
}

// Info writes the LINFO level log.
func Info(args ...interface{}) {
	if logger != nil {
		logger.Info(args...)
	}
}

// Infof writes the formated LINFO level log.
func Infof(format string, args ...interface{}) {
	if logger != nil {
		logger.Infof(format, args...)
	}
}

// Level returns current logger Level.
func Level() unilogger.Level {
	return currentLevel
}

// Logger returns default logger.
func Logger() unilogger.LeveledLogger {
	return logger
}

// Panic writes and panics the log.
func Panic(args ...interface{}) {
	if logger != nil {
		logger.Panic(args...)
	} else {
		panic(fmt.Sprint(args...))
	}
}

// Panicf writes and panics formatted log.
func Panicf(format string, args ...interface{}) {
	if logger != nil {
		logger.Panicf(format, args...)
	} else {
		panic(fmt.Sprintf(format, args...))
	}
}

// SetLevel sets the level if possible for the logger file.
func SetLevel(level unilogger.Level) error {
	if level == LUNKNOWN {
		return errors.NewDet(class.CommonLoggerUnknownLevel, "can't set unknown logger level. provided level is not valid")
	}

	// level is already the same
	if level == currentLevel {
		Debug3f("Current level the same as the provided: '%s' level.", level)
		return nil
	}

	currentLevel = level
	if logger == nil {
		return nil
	}

	lvl, ok := logger.(unilogger.LevelSetter)
	if !ok {
		return errors.NewDet(class.CommonLoggerNotImplement, "logger doesn't implement LevelSetter interface")
	}

	lvl.SetLevel(currentLevel)
	return nil
}

// SetLogger sets the 'log' as the current logger.
func SetLogger(log unilogger.LeveledLogger) {
	logger = log

	depth, ok := log.(unilogger.OutputDepthGetter)
	if ok {
		setter, ok := log.(unilogger.OutputDepthSetter)
		if ok {
			setter.SetOutputDepth(depth.GetOutputDepth() + 1)
		}
	}

	if lvlSetter, ok := log.(unilogger.LevelSetter); ok {
		lvlSetter.SetLevel(currentLevel)
	}

	Debug("New logger set with level: %s", currentLevel.String())

	debugLeveled, isDebugLeveled = log.(unilogger.DebugLeveledLogger)
	subLogger, isSubLogger := log.(unilogger.SubLogger)
	for _, m := range modules {
		if m.logger == nil && isSubLogger {
			m.logger = subLogger.SubLogger()
			m.initializeLogger()
		}
		m.SetLevel(currentLevel)
	}
}

// SetModulesLevel sets the 'level' for all modules.
func SetModulesLevel(level unilogger.Level) error {
	if level == LUNKNOWN {
		return errors.NewDet(class.CommonLoggerUnknownLevel, "can't set unknown logger level. provided level is not valid")
	}
	for _, module := range modules {
		module.SetLevel(level)
	}
	return nil
}

// New creates new unilogger.BasicLogger that writes to provided 'out' io.Writer
// with specific 'prefix' and provided 'flags'.
func New(out io.Writer, prefix string, flags int) {
	basic := unilogger.NewBasicLogger(out, prefix, flags)
	basic.SetOutputDepth(4)
	SetLogger(basic)
}

// Warning writes the warning level log.
func Warning(args ...interface{}) {
	if logger != nil {
		logger.Warning(args...)
	}
}

// Warningf writes the formated warning level log.
func Warningf(format string, args ...interface{}) {
	if logger != nil {
		logger.Warningf(format, args...)
	}
}
