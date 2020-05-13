package log

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMessage tests the Message methods.
func TestMessage(t *testing.T) {
	t.Run("LogMessage", func(t *testing.T) {
		t.Run("NilFmt", func(t *testing.T) {
			message := &Message{
				id:    1,
				level: LevelInfo,
				fmt:   nil,
				args: []interface{}{
					"Some argument", "SecondArgument",
				},
			}
			msgStr := message.Message()
			assert.Equal(t, fmt.Sprint(message.args...), msgStr)
		})
		t.Run("NonNilFmt", func(t *testing.T) {
			format := "%s, %s"
			message := &Message{
				id:    2,
				level: LevelDebug,
				fmt:   &format,
				args: []interface{}{
					"first", "second",
				},
			}
			msgStr := message.Message()
			assert.Equal(t, fmt.Sprintf(*message.fmt, message.args...), msgStr)

			str := message.String()
			assert.Equal(t, fmt.Sprintf("%s|%04x: %s", message.level, message.id, message.getMessage()), str)
		})
	})
}

// TestBasicLogger tests the basic logger functions.
func TestBasicLogger(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewBasicLogger(&buf, "", 0)
		assert.NotNil(t, logger)
		assert.IsType(t, &BasicLogger{}, logger)
		assert.NotNil(t, logger.stdLogger)
		assert.Equal(t, LevelInfo, logger.level)

		t.Run("SetLevel", func(t *testing.T) {
			logger.SetLevel(LevelError)
			assert.Equal(t, LevelError, logger.level)
		})

		args := []interface{}{"First", "Second"}
		format := "%s-%s"
		msg := prepareMessage(logSequenceID, LevelDebug, nil, args...)
		msgFmt := prepareMessage(logSequenceID, LevelDebug, &format, args...)

		t.Run("ExtendedLeveledLogger", func(t *testing.T) {
			assert.Implements(t, (*LeveledLogger)(nil), logger)

			logger.SetLevel(LevelDebug)

			buf.Reset()
			logger.Debug(args...)
			msg.id = logSequenceID
			assert.Equal(t, fmtMsg(msg), buf.String())

			buf.Reset()
			logger.Debugf(format, args...)
			msgFmt.id = logSequenceID
			assert.Equal(t, fmtMsg(msgFmt), buf.String())
		})

		t.Run("LevelAbove", func(t *testing.T) {
			logger.SetLevel(LevelInfo)
			buf.Reset()
			logger.Debug(args...)

			assert.Empty(t, buf.String())
		})

		t.Run("Info", func(t *testing.T) {
			logger.SetLevel(LevelInfo)
			buf.Reset()
			logger.Info(args...)
			msg.level = LevelInfo
			msg.id = logSequenceID
			assert.Equal(t, fmtMsg(msg), buf.String())

			buf.Reset()
			logger.Infof(format, args...)
			msgFmt.level = LevelInfo
			msgFmt.id = logSequenceID
			assert.Equal(t, fmtMsg(msgFmt), buf.String())
		})

		t.Run("Warning", func(t *testing.T) {
			logger.SetLevel(LevelWarning)
			buf.Reset()
			logger.Warning(args...)
			msg.level = LevelWarning
			msg.id = logSequenceID
			assert.Equal(t, fmtMsg(msg), buf.String())

			buf.Reset()
			logger.Warningf(format, args...)
			msgFmt.level = LevelWarning
			msgFmt.id = logSequenceID
			assert.Equal(t, fmtMsg(msgFmt), buf.String())
		})

		t.Run("Error", func(t *testing.T) {
			logger.SetLevel(LevelError)
			buf.Reset()
			logger.Error(args...)
			msg.level = LevelError
			msg.id = logSequenceID
			assert.Equal(t, fmtMsg(msg), buf.String())

			buf.Reset()
			logger.Errorf(format, args...)
			msgFmt.level = LevelError
			msgFmt.id = logSequenceID
			assert.Equal(t, fmtMsg(msgFmt), buf.String())
		})

		t.Run("Panic", func(t *testing.T) {
			logger.SetLevel(LevelCritical)
			buf.Reset()

			assert.Panics(t, func() { logger.Panic(args...) })
			msg.level = LevelCritical
			msg.id = logSequenceID
			assert.Equal(t, fmtMsg(msg), buf.String())

			buf.Reset()
			msgFmt.level = LevelCritical
			assert.Panics(t, func() { logger.Panicf(format, args...) })
			msgFmt.id = logSequenceID
			assert.Equal(t, fmtMsg(msgFmt), buf.String())
		})
	})
}

func prepareMessage(id uint64, level Level, fmt *string, args ...interface{}) *Message {
	return &Message{id: id, level: level, fmt: fmt, args: args}
}

func fmtMsg(msg *Message) string {
	return fmt.Sprintf("%s\n", msg.String())
}
