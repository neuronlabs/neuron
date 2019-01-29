package log

import (
	"github.com/kucjac/uni-logger"
	"github.com/pkg/errors"
	"io"
	"log"
	"os"
)

var logger unilogger.LeveledLogger

// Logger returns default logger
func Logger() unilogger.LeveledLogger {
	return logger
}

// SetLevel sets the level if possible for the logger file
func SetLevel(level unilogger.Level) error {
	if logger == nil {
		Default()
	}

	lvl, ok := logger.(unilogger.LevelSetter)
	if !ok {
		return errors.New("logger doesn't implement LevelSetter interface")
	}

	lvl.SetLevel(level)
	return nil
}

// SetLogger sets the logger for the provided 'log' logger
func SetLogger(log unilogger.LeveledLogger) {
	logger = log

	depth, ok := log.(unilogger.OutputDepthGetter)
	if ok {
		setter, ok := log.(unilogger.OutputDepthSetter)
		if ok {
			setter.SetOutputDepth(depth.GetOutputDepth() + 1)
		}
	}
}

// New creates new logger on the base of the provided values
func New(out io.Writer, prefix string, flags int) {
	basic := unilogger.NewBasicLogger(out, prefix, flags)
	basic.SetOutputDepth(4)
	logger = basic
}

// Default creates default logger
func Default() {
	basic := unilogger.NewBasicLogger(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile)
	basic.SetOutputDepth(4)
	logger = basic
}

// Debugf writes the formated debug log
func Debugf(format string, args ...interface{}) {
	if logger != nil {
		logger.Debugf(format, args...)
	}
}

// Infof writes the formated info log
func Infof(format string, args ...interface{}) {
	if logger != nil {
		logger.Infof(format, args...)
	}
}

// Warningf writes the formated warning log
func Warningf(format string, args ...interface{}) {
	if logger != nil {
		logger.Warningf(format, args...)
	}
}

// Errorf writes the formated error log
func Errorf(format string, args ...interface{}) {
	if logger != nil {
		logger.Errorf(format, args...)
	}
}

// Fatalf writes the formated fatal log
func Fatalf(format string, args ...interface{}) {
	if logger != nil {
		logger.Fatalf(format, args...)
	}
}

// Panicf writes the formated panic log
func Panicf(format string, args ...interface{}) {
	if logger != nil {
		logger.Panicf(format, args...)
	}
}

// Debug writes the  debug log
func Debug(args ...interface{}) {
	if logger != nil {
		logger.Debug(args...)
	}
}

// Info writes the info log
func Info(args ...interface{}) {
	if logger != nil {
		logger.Info(args...)
	}
}

// Warning writes the warning log
func Warning(args ...interface{}) {
	if logger != nil {
		logger.Warning(args...)
	}
}

// Error writes the error log
func Error(args ...interface{}) {
	if logger != nil {
		logger.Error(args...)
	}
}

// Fatal writes the fatal log
func Fatal(args ...interface{}) {
	if logger != nil {
		logger.Fatal(args...)
	}
}

// Panic writes the panic log
func Panic(args ...interface{}) {
	if logger != nil {
		logger.Panic(args...)
	}
}
