package log

import (
	"github.com/neuronlabs/uni-logger"
)

// ModuleLogger is the logger used for getting the specific modules.
type ModuleLogger struct {
	Name           string
	logger         unilogger.LeveledLogger
	isDebugLeveled bool
	debugLeveled   unilogger.DebugLeveledLogger
}

// NewModuleLogger creates new module logger for given 'name' of the module and an optional 'logger'.
func NewModuleLogger(name string, moduleLogger ...unilogger.LeveledLogger) *ModuleLogger {
	mLogger := &ModuleLogger{
		Name: name,
	}

	if len(moduleLogger) > 0 {
		mLogger.logger = moduleLogger[0]
		mLogger.debugLeveled, mLogger.isDebugLeveled = moduleLogger[0].(unilogger.DebugLeveledLogger)

		depthGetter, isDepthGetter := moduleLogger[0].(unilogger.OutputDepthGetter)
		if isDepthGetter {
			depthSetter, isDepthSetter := moduleLogger[0].(unilogger.OutputDepthSetter)
			if isDepthSetter {
				depthSetter.SetOutputDepth(depthGetter.GetOutputDepth() + 1)
			}
		}
	}
	return mLogger
}

func (m *ModuleLogger) log() unilogger.LeveledLogger {
	if m.logger != nil {
		return m.logger
	}
	return logger
}

// Debug3f writes the formated debug3 log.
func (m *ModuleLogger) Debug3f(format string, args ...interface{}) {
	format = m.name() + " " + format
	if m.logger != nil {
		if !m.isDebugLeveled {
			m.logger.Debugf(format, args...)
		} else {
			m.debugLeveled.Debug3f(format, args...)
		}
	} else {
		Debug3f(format, args...)
	}
}

// Debug2f writes the formated debug2 log.
func (m *ModuleLogger) Debug2f(format string, args ...interface{}) {
	format = m.name() + " " + format
	if m.logger != nil {
		if !m.isDebugLeveled {
			m.logger.Debugf(format, args...)
		} else {
			m.debugLeveled.Debug2f(format, args...)
		}
	} else {
		Debug2f(format, args...)
	}
}

// Debugf writes the formated debug log.
func (m *ModuleLogger) Debugf(format string, args ...interface{}) {
	format = m.name() + " " + format
	if m.logger != nil {
		m.logger.Debugf(format, args...)
	} else {
		Debugf(format, args...)
	}
}

// Infof writes the formated info log.
func (m *ModuleLogger) Infof(format string, args ...interface{}) {
	format = m.name() + " " + format
	if m.logger != nil {
		m.logger.Infof(format, args...)
	} else {
		Infof(format, args...)
	}
}

// Warningf writes the formated warning log.
func (m *ModuleLogger) Warningf(format string, args ...interface{}) {
	format = m.name() + " " + format
	if m.logger != nil {
		m.logger.Warningf(format, args...)
	} else {
		Warningf(format, args...)
	}
}

// Errorf writes the formated error log.
func (m *ModuleLogger) Errorf(format string, args ...interface{}) {
	format = m.name() + " " + format
	if m.logger != nil {
		m.logger.Errorf(format, args...)
	} else {
		Errorf(format, args...)
	}
}

// Fatalf writes the formated fatal log.
func (m *ModuleLogger) Fatalf(format string, args ...interface{}) {
	format = m.name() + " " + format
	if m.logger != nil {
		m.logger.Fatalf(format, args...)
	} else {
		Fatalf(format, args...)
	}
}

// Panicf writes the formated panic log.
func (m *ModuleLogger) Panicf(format string, args ...interface{}) {
	format = m.name() + " " + format
	if m.logger != nil {
		m.logger.Panicf(format, args...)
	} else {
		Panicf(format, args...)
	}
}

// Debug3 writes the debug3 level log.
func (m *ModuleLogger) Debug3(args ...interface{}) {
	args = append([]interface{}{m.name(), " "}, args...)
	if m.logger != nil {
		if !m.isDebugLeveled {
			m.logger.Debug(args...)
		} else {
			m.debugLeveled.Debug3(args...)
		}
	} else {
		Debug3(args...)
	}
}

// Debug2 writes the debug2 level log.
func (m *ModuleLogger) Debug2(args ...interface{}) {
	args = append([]interface{}{m.name(), " "}, args...)
	if m.logger != nil {
		if !m.isDebugLeveled {
			m.logger.Debug(args...)
		} else {
			m.debugLeveled.Debug2(args...)
		}
	} else {
		Debug2(args...)
	}
}

// Debug writes the debug level log.
func (m *ModuleLogger) Debug(args ...interface{}) {
	args = append([]interface{}{m.name(), " "}, args...)
	if m.logger != nil {
		m.logger.Debug(args...)
	} else {
		Debug(args...)
	}
}

// Info writes the info level log.
func (m *ModuleLogger) Info(args ...interface{}) {
	args = append([]interface{}{m.name(), " "}, args...)
	if m.logger != nil {
		m.logger.Info(args...)
	} else {
		Info(args...)
	}
}

// Warning writes the warning level log.
func (m *ModuleLogger) Warning(args ...interface{}) {
	args = append([]interface{}{m.name(), " "}, args...)
	if m.logger != nil {
		m.logger.Warning(args...)
	} else {
		Warning(args...)
	}
}

// Error writes the error level log.
func (m *ModuleLogger) Error(args ...interface{}) {
	args = append([]interface{}{m.name(), " "}, args...)
	if m.logger != nil {
		m.logger.Error(args...)
	} else {
		Error(args...)
	}
}

// Fatal writes the fatal level log.
func (m *ModuleLogger) Fatal(args ...interface{}) {
	args = append([]interface{}{m.name(), " "}, args...)
	if m.logger != nil {
		m.logger.Fatal(args...)
	} else {
		Fatal(args...)
	}
}

// Panic writes the panic level log.
func (m *ModuleLogger) Panic(args ...interface{}) {
	args = append([]interface{}{m.name(), " "}, args...)
	if m.logger != nil {
		m.logger.Panic(args...)
	} else {
		Panic(args...)
	}
}

func (m *ModuleLogger) name() string {
	return "[" + m.Name + "]"
}
