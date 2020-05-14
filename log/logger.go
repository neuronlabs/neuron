package log

import (
	"io"
	"log"
	"os"

	"github.com/neuronlabs/neuron/errors"
)

var (
	defaultLogger  LeveledLogger = dummyLogger{}
	currentLevel                 = LevelInfo
	debugLeveled   DebugLeveledLogger
	isDebugLeveled bool
)

// Debug writes the LevelDebug level log.
func Debug(args ...interface{}) {
	defaultLogger.Debug(args...)
}

// Debug2 writes the LevelDebug2 level logs.
func Debug2(args ...interface{}) {
	if !isDebugLeveled {
		defaultLogger.Debug(args...)
	} else {
		debugLeveled.Debug2(args...)
	}
}

// Debug2f writes the formatted LevelDebug2 level log.
func Debug2f(format string, args ...interface{}) {
	if !isDebugLeveled {
		defaultLogger.Debugf(format, args...)
	} else {
		debugLeveled.Debug2f(format, args...)
	}
}

// Debug3 writes the LevelDebug3 level logs.
func Debug3(args ...interface{}) {
	if !isDebugLeveled {
		defaultLogger.Debug(args...)
	} else {
		debugLeveled.Debug3(args...)
	}
}

// Debug3f writes the formatted LevelDebug2 level log.
func Debug3f(format string, args ...interface{}) {
	if !isDebugLeveled {
		defaultLogger.Debugf(format, args...)
	} else {
		debugLeveled.Debug3f(format, args...)
	}
}

// Debugf writes the formatted LevelDebug level log.
func Debugf(format string, args ...interface{}) {
	defaultLogger.Debugf(format, args...)
}

// Error writes the LevelError level log.
func Error(args ...interface{}) {
	defaultLogger.Error(args...)
}

// Errorf writes the formatted LevelError level log.
func Errorf(format string, args ...interface{}) {
	defaultLogger.Errorf(format, args...)
}

// Fatal writes the fatal - LevelCritical level log.
func Fatal(args ...interface{}) {
	defaultLogger.Fatal(args...)
}

// Fatalf writes the formatted fatal - LevelCritical level log.
func Fatalf(format string, args ...interface{}) {
	defaultLogger.Fatalf(format, args...)
}

// Info writes the LevelInfo level log.
func Info(args ...interface{}) {
	defaultLogger.Info(args...)
}

// Infof writes the formatted LevelInfo level log.
func Infof(format string, args ...interface{}) {
	defaultLogger.Infof(format, args...)
}

// CurrentLevel returns current logger Level.
func CurrentLevel() Level {
	return currentLevel
}

// Logger returns default logger.
func Logger() LeveledLogger {
	return defaultLogger
}

// Panic writes and panics the log.
func Panic(args ...interface{}) {
	defaultLogger.Panic(args...)
}

// Panicf writes and panics formatted log.
func Panicf(format string, args ...interface{}) {
	defaultLogger.Panicf(format, args...)
}

// SetLevel sets the level if possible for the logger file.
func SetLevel(level Level) error {
	if level == LevelUnknown {
		return errors.NewDet(ClassInvalidLogger, "can't set unknown logger level. provided level is not valid")
	}

	// level is already the same
	if level == currentLevel {
		Debug3f("Current level is the same as provided: '%s' level.", level)
		return nil
	}

	currentLevel = level
	if defaultLogger == nil {
		return nil
	}

	lvl, ok := defaultLogger.(LevelSetter)
	if !ok {
		return errors.NewDet(ClassInvalidLogger, "logger doesn't implement LevelSetter interface")
	}

	lvl.SetLevel(currentLevel)
	return nil
}

// SetLogger sets the 'log' as the current logger.
func SetLogger(log LeveledLogger) {
	defaultLogger = log
	depthGetter, ok := log.(OutputDepthGetter)
	if ok {
		depth := depthGetter.GetOutputDepth()
		if depth != 4 {
			setter, ok := log.(OutputDepthSetter)
			if ok {
				setter.SetOutputDepth(4)
			}
		}
	}

	if lvlSetter, ok := log.(LevelSetter); ok {
		lvlSetter.SetLevel(currentLevel)
	}
	Debug("New logger set with level: %s", currentLevel.String())

	debugLeveled, isDebugLeveled = log.(DebugLeveledLogger)
	subLogger, isSubLogger := log.(SubLogger)
	for _, m := range modules {
		// if the module loggers are not initialized - create them now.
		if m.logger == nil && isSubLogger {
			m.logger = subLogger.SubLogger()
			m.initializeLogger()
		}
		m.SetLevel(currentLevel)
	}
}

// SetModulesLevel sets the 'level' for all modules.
func SetModulesLevel(level Level) error {
	if level == LevelUnknown {
		return errors.NewDet(ClassInvalidLogger, "can't set unknown logger level. provided level is not valid")
	}
	for _, module := range modules {
		module.SetLevel(level)
	}
	return nil
}

// New creates new basic logger that writes to provided 'out' io.Writer
// with specific 'prefix' and provided 'flags'.
func New(out io.Writer, prefix string, flags int) {
	basic := NewBasicLogger(out, prefix, flags)
	basic.SetOutputDepth(4)
	SetLogger(basic)
}

// NewDefault creates new basic logger without prefix with standard flags.
func NewDefault() {
	basic := NewBasicLogger(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile)
	basic.SetOutputDepth(4)
	basic.SetLevel(LevelInfo)
	SetLogger(basic)
}

// Warning writes the warning level log.
func Warning(args ...interface{}) {
	defaultLogger.Warning(args...)
}

// Warningf writes the formatted warning level log.
func Warningf(format string, args ...interface{}) {
	defaultLogger.Warningf(format, args...)
}
