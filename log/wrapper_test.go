package log

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type stdlogger struct{}

func (s *stdlogger) Print(args ...interface{})                 {}
func (s *stdlogger) Printf(format string, args ...interface{}) {}
func (s *stdlogger) Println(args ...interface{})               {}
func (s *stdlogger) Panic(args ...interface{})                 {}
func (s *stdlogger) Panicf(format string, args ...interface{}) {}
func (s *stdlogger) Panicln(args ...interface{})               {}
func (s *stdlogger) Fatal(args ...interface{})                 {}
func (s *stdlogger) Fatalf(format string, args ...interface{}) {}
func (s *stdlogger) Fatalln(args ...interface{})               {}

type leveledLogger struct{}

func (l *leveledLogger) Debugf(format string, args ...interface{})   {}
func (l *leveledLogger) Infof(format string, args ...interface{})    {}
func (l *leveledLogger) Warningf(format string, args ...interface{}) {}
func (l *leveledLogger) Errorf(format string, args ...interface{})   {}
func (l *leveledLogger) Fatalf(format string, args ...interface{})   {}
func (l *leveledLogger) Panicf(format string, args ...interface{})   {}
func (l *leveledLogger) Debug(args ...interface{})                   {}
func (l *leveledLogger) Info(args ...interface{})                    {}
func (l *leveledLogger) Warning(args ...interface{})                 {}
func (l *leveledLogger) Error(args ...interface{})                   {}
func (l *leveledLogger) Fatal(args ...interface{})                   {}
func (l *leveledLogger) Panic(args ...interface{})                   {}

type shortLeveledLogger struct{}

func (c *shortLeveledLogger) Debugf(format string, args ...interface{}) {}
func (c *shortLeveledLogger) Infof(format string, args ...interface{})  {}
func (c *shortLeveledLogger) Warnf(format string, args ...interface{})  {}
func (c *shortLeveledLogger) Errorf(format string, args ...interface{}) {}
func (c *shortLeveledLogger) Fatalf(format string, args ...interface{}) {}
func (c *shortLeveledLogger) Panicf(format string, args ...interface{}) {}
func (c *shortLeveledLogger) Debug(args ...interface{})                 {}
func (c *shortLeveledLogger) Info(args ...interface{})                  {}
func (c *shortLeveledLogger) Warn(args ...interface{})                  {}
func (c *shortLeveledLogger) Error(args ...interface{})                 {}
func (c *shortLeveledLogger) Fatal(args ...interface{})                 {}
func (c *shortLeveledLogger) Panic(args ...interface{})                 {}

type extendedLogger struct {
	leveledLogger
}

func (e *extendedLogger) Print(args ...interface{})                 {}
func (e *extendedLogger) Printf(format string, args ...interface{}) {}
func (e *extendedLogger) Println(args ...interface{})               {}
func (e *extendedLogger) Debugln(args ...interface{})               {}
func (e *extendedLogger) Infoln(args ...interface{})                {}
func (e *extendedLogger) Warningln(args ...interface{})             {}
func (e *extendedLogger) Errorln(args ...interface{})               {}
func (e *extendedLogger) Fatalln(args ...interface{})               {}
func (e *extendedLogger) Panicln(args ...interface{})               {}

type nonLogger struct{}

// TestNewLoggerWrapper tests NewLoggerWrapper function.
func TestNewLoggerWrapper(t *testing.T) {
	loggers := []interface{}{&stdlogger{}, &leveledLogger{}, &shortLeveledLogger{}, &extendedLogger{}}

	t.Run("NewWrapper", func(t *testing.T) {
		for _, logger := range loggers {
			wrapper, err := NewLoggerWrapper(logger)
			assert.IsType(t, &Wrapper{}, wrapper)
			assert.NoError(t, err)
			wrapper = MustGetLoggerWrapper(logger)
			assert.IsType(t, &Wrapper{}, wrapper)
		}

		args := []interface{}{}
		format := "some format"
		for _, logger := range loggers {
			wrapper := MustGetLoggerWrapper(logger)
			wrapper.Print(args)
			wrapper.Printf(format, args)
			wrapper.Println(args)

			wrapper.Debug(args)
			wrapper.Debugf(format, args...)
			wrapper.Debugln(args)

			wrapper.Info(args)
			wrapper.Infof(format, args...)
			wrapper.Infoln(args)

			wrapper.Warning(args)
			wrapper.Warningf(format, args...)
			wrapper.Warningln(args)

			wrapper.Error(args)
			wrapper.Errorf(format, args)
			wrapper.Errorln(args)

			wrapper.Fatal(args)
			wrapper.Fatalf(format, args)
			wrapper.Fatalln(args)

			wrapper.Panic(args)
			wrapper.Panicf(format, args)
			wrapper.Panicln(args)
		}
	})

	t.Run("NotImplement", func(t *testing.T) {
		unknownLogger := nonLogger{}
		wrapper, err := NewLoggerWrapper(unknownLogger)
		assert.Error(t, err)
		assert.Nil(t, wrapper)
		assert.Panics(t, func() { MustGetLoggerWrapper(unknownLogger) })
	})
}

// TestBuildLeveled tests the buildLeveled function.
func TestBuildLeveled(t *testing.T) {
	level := LevelDebug
	format := "some format"
	arguments := []interface{}{"First", "Second"}

	// Providing nil format should add level as first argument to args
	t.Run("NilFormat", func(t *testing.T) {
		args := buildLeveled(level, nil, arguments...)
		assert.Equal(t, fmt.Sprintf("%s: ", level), args[0])
	})

	t.Run("Formatted", func(t *testing.T) {
		thisFormat := format
		buildLeveled(level, &thisFormat, arguments...)
		assert.NotEqual(t, format, thisFormat)
	})
}
