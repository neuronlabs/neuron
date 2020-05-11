package log

// SubLogger interface that creates and returns new sub logger.
type SubLogger interface {
	SubLogger() LeveledLogger
}

// LevelSetter is the interface that allows to set the logging level.
type LevelSetter interface {
	SetLevel(level Level)
}

// LevelGetter is the interface used to get current logger level.
type LevelGetter interface {
	GetLevel() Level
}

// OutputDepthSetter is the interface that sets the output depth for the logging interface.
type OutputDepthSetter interface {
	SetOutputDepth(depth int)
}

// OutputDepthGetter is the interface that gets the output get.
type OutputDepthGetter interface {
	GetOutputDepth() int
}

// StdLogger is the logger interface for standard log library.
type StdLogger interface {
	Print(args ...interface{})
	Printf(format string, args ...interface{})
	Println(args ...interface{})

	Panic(args ...interface{})
	Panicf(format string, args ...interface{})
	Panicln(args ...interface{})

	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Fatalln(args ...interface{})
}

// LeveledLogger is a logger that uses basic logging levels.
type LeveledLogger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warningf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Panicf(format string, args ...interface{})

	Debug(args ...interface{})
	Info(args ...interface{})
	Warning(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Panic(args ...interface{})
}

// DebugLeveledLogger is Leveled Logger with the debug2 and debug3 levels.
type DebugLeveledLogger interface {
	LeveledLogger
	Debug2(args ...interface{})
	Debug3(args ...interface{})

	Debug2f(format string, args ...interface{})
	Debug3f(format string, args ...interface{})
}

// ShortLeveledLogger is a logger that uses basic logging levels.
// with short name for Warn.
type ShortLeveledLogger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Panicf(format string, args ...interface{})

	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Panic(args ...interface{})
}

// ExtendedLeveledLogger adds distinction between Leveled methods that starts new or not.
// i.e.: 'Debugln' and 'Debug'.
// It also adds all Print's methods.
type ExtendedLeveledLogger interface {
	Print(args ...interface{})
	Printf(format string, args ...interface{})
	Println(args ...interface{})

	Debug3f(format string, args ...interface{})
	Debug2f(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warningf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Panicf(format string, args ...interface{})

	Debug3(format string, args ...interface{})
	Debug2(format string, args ...interface{})
	Debug(args ...interface{})
	Info(args ...interface{})
	Warning(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Panic(args ...interface{})

	Debug3ln(args ...interface{})
	Debug2ln(args ...interface{})
	Debugln(args ...interface{})
	Infoln(args ...interface{})
	Warningln(args ...interface{})
	Errorln(args ...interface{})
	Fatalln(args ...interface{})
	Panicln(args ...interface{})
}
