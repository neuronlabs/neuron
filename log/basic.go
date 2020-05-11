package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync/atomic"
)

var logSequenceID = uint64(0)

/**

Levels

*/

// Level defines a logging level used in BasicLogger.
type Level int

// Following levels are supported in BasicLogger.
const (
	DEBUG3 Level = iota
	DEBUG2
	DEBUG
	INFO
	WARNING
	ERROR
	CRITICAL
	UNKNOWN
)

// IsAllowed checks if the 'other' Level is allowed to be used in compare with 'l' Level.
func (l Level) IsAllowed(other Level) bool {
	return other >= l
}

var levelNames = map[Level]string{
	DEBUG3:   "DEBUG3",
	DEBUG2:   "DEBUG2",
	DEBUG:    "DEBUG",
	INFO:     "INFO",
	WARNING:  "WARNING",
	ERROR:    "ERROR",
	CRITICAL: "CRITICAL",
}

func (l Level) String() string {
	return levelNames[l]
}

// ParseLevel parses level from string.
func ParseLevel(level string) Level {
	level = strings.ToUpper(level)

	for l, name := range levelNames {
		if name == level {
			return l
		}
	}
	return UNKNOWN
}

/**

Message

*/

// Message is a basic logging record structure used in BasicLogger.
type Message struct {
	id      uint64
	level   Level
	fmt     *string
	message *string
	args    []interface{}
}

// Message prepares the string message based on the format and args private fields
// of the message.
func (m *Message) Message() string {
	return m.getMessage()
}

func (m *Message) getMessage() string {
	if m.message == nil {
		var msg string
		if m.fmt == nil {
			//println etc.
			msg = fmt.Sprint(m.args...)
		} else {
			msg = fmt.Sprintf(*m.fmt, m.args...)
		}
		m.message = &msg
	}
	return *m.message
}

// String returns string that concantates:
// id hash - 4 digits|time formatted in RFC339|level|message.
// Implements fmt.Stringer interface.
func (m *Message) String() string {
	msg := fmt.Sprintf("%s|%04x: %s", m.level, m.id, m.getMessage())
	return msg
}

/**

BasicLogger

*/

// BasicLogger is simple leveled logger that implements DebugLeveledLogger interface.
// It uses 7 log levels:
//  # DEBUG3
//  # DEBUG2
//	# DEBUG
//	# INFO
//	# WARNING
//	# ERROR
//	# CRITICAL
// By default INFO level is used. It may be reset using SetLevel() method.
// It allows to filter the logs by given level.
// I.e. Having BasicLogger with level Set to WARNING, then there would be
// no DEBUG and INFO logs (the hierarchy goes up only).
type BasicLogger struct {
	stdLogger   *log.Logger
	level       Level
	outputDepth int
}

var _ DebugLeveledLogger = &BasicLogger{}

// NewBasicLogger creates new BasicLogger that shares common sequence id.
// By default it uses DEBUG level. It can be changed later using SetLevel() method.
// BasicLogger uses standard library *log.Logger for logging purpose.
// The arguments used in this function are described in log.New() method.
func NewBasicLogger(out io.Writer, prefix string, flags int) *BasicLogger {
	logger := &BasicLogger{
		stdLogger:   log.New(out, prefix, flags),
		level:       INFO,
		outputDepth: 3,
	}
	return logger
}

var _ SubLogger = &BasicLogger{}

// SubLogger creates new sublogger for given logger.
func (l *BasicLogger) SubLogger() LeveledLogger {
	sub := &BasicLogger{
		stdLogger:   l.stdLogger,
		level:       l.level,
		outputDepth: 4,
	}
	return sub
}

var _ LevelSetter = &BasicLogger{}

// SetLevel sets the level of logging for given Logger.
func (l *BasicLogger) SetLevel(level Level) {
	l.Debugf("Setting log level to: '%s'", level)
	l.level = level
}

// SetOutputDepth set sthe output depth of the basic logger
// the output depth is the standard logger function depths
func (l *BasicLogger) SetOutputDepth(depth int) {
	l.outputDepth = depth
}

// GetOutputDepth gets the output depth
func (l *BasicLogger) GetOutputDepth() int {
	return l.outputDepth
}

// GetLevel gets current logger level.
func (l *BasicLogger) GetLevel() Level {
	return l.level
}

// Debug3 logs a message with DEBUG level.
func (l *BasicLogger) Debug3(args ...interface{}) {
	l.log(DEBUG3, nil, args...)
}

// Debug3f logs a formatted message with DEBUG level.
func (l *BasicLogger) Debug3f(format string, args ...interface{}) {
	l.log(DEBUG3, &format, args...)
}

// Debug2 logs a message with DEBUG level.
func (l *BasicLogger) Debug2(args ...interface{}) {
	l.log(DEBUG2, nil, args...)
}

// Debug2f logs a formatted message with DEBUG level.
func (l *BasicLogger) Debug2f(format string, args ...interface{}) {
	l.log(DEBUG2, &format, args...)
}

// Debug logs a message with DEBUG level.
func (l *BasicLogger) Debug(args ...interface{}) {
	l.log(DEBUG, nil, args...)
}

// Debugf logs a formatted message with DEBUG level.
func (l *BasicLogger) Debugf(format string, args ...interface{}) {
	l.log(DEBUG, &format, args...)
}

// Info logs a message with INFO level.
func (l *BasicLogger) Info(args ...interface{}) {
	l.log(INFO, nil, args...)
}

// Infof logs a formatted message with INFO level.
func (l *BasicLogger) Infof(format string, args ...interface{}) {
	l.log(INFO, &format, args...)
}

// Print logs a message. Arguments are handled in a log.Print manner.
func (l *BasicLogger) Print(args ...interface{}) {
	l.log(INFO, nil, args...)
}

// Printf logs a formatted message. Arguments are handled in a log.Printf manner.
func (l *BasicLogger) Printf(format string, args ...interface{}) {
	l.log(INFO, &format, args...)
}

// Warning logs a message with WARNING level. Arguments are handled in a log.Print manner.
func (l *BasicLogger) Warning(args ...interface{}) {
	l.log(WARNING, nil, args...)
}

// Warningf logs a formatted message with WARNING level. Arguments are handled in a log.Printf manner.
func (l *BasicLogger) Warningf(format string, args ...interface{}) {
	l.log(WARNING, &format, args...)
}

// Error logs a message with ERROR level. Arguments are handled in a log.Print manner.
func (l *BasicLogger) Error(args ...interface{}) {
	l.log(ERROR, nil, args...)
}

// Errorf logs a formatted message with ERROR level. Arguments are handled in a log.Printf manner.
func (l *BasicLogger) Errorf(format string, args ...interface{}) {
	l.log(ERROR, &format, args...)
}

// Fatal logs a message with CRITICAL level. Afterwards the function execute os.Exit(1).
// Arguments are handled in a log.Print manner.
func (l *BasicLogger) Fatal(args ...interface{}) {
	l.log(CRITICAL, nil, args...)
	os.Exit(1)
}

// Fatalf logs a formatted message with CRITICAL level. Afterwards the function execute os.Exit(1).
// Arguments are handled in a log.Printf manner.
func (l *BasicLogger) Fatalf(format string, args ...interface{}) {
	l.log(CRITICAL, &format, args...)
	os.Exit(1)
}

// Panic logs a message with CRITICAL level. Afterwards the function panics with given message.
// Arguments are handled in a log.Print manner.
func (l *BasicLogger) Panic(args ...interface{}) {
	l.log(CRITICAL, nil, args...)
	panic(fmt.Sprint(args...))
}

// Panicf logs a formatted message with CRITICAL level. Afterwards the function panics with given
// formatted message. Arguments are handled in a log.Printf manner.
func (l *BasicLogger) Panicf(format string, args ...interface{}) {
	l.log(CRITICAL, &format, args...)
	panic(fmt.Sprintf(format, args...))
}

/**

PRIVATE

*/

func (l *BasicLogger) log(level Level, format *string, args ...interface{}) {
	if !l.isLevelEnabled(level) {
		return
	}
	msg := &Message{
		id:    atomic.AddUint64(&logSequenceID, 1),
		level: level,
		fmt:   format,
		args:  args,
	}

	l.stdLogger.Output(l.outputDepth, msg.String())
}

func (l *BasicLogger) isLevelEnabled(level Level) bool {
	return level >= l.level
}
