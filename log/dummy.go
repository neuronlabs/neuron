package log

import (
	"fmt"
	"os"
)

// Compile time check for dummyLogger to implement ExtendedLeveledLogger.
var _ ExtendedLeveledLogger = dummyLogger{}

type dummyLogger struct{}

func (d dummyLogger) Print(...interface{})            {}
func (d dummyLogger) Printf(string, ...interface{})   {}
func (d dummyLogger) Println(...interface{})          {}
func (d dummyLogger) Debug3f(string, ...interface{})  {}
func (d dummyLogger) Debug2f(string, ...interface{})  {}
func (d dummyLogger) Debugf(string, ...interface{})   {}
func (d dummyLogger) Infof(string, ...interface{})    {}
func (d dummyLogger) Warningf(string, ...interface{}) {}
func (d dummyLogger) Errorf(string, ...interface{})   {}
func (d dummyLogger) Fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}
func (d dummyLogger) Panicf(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args...))
}
func (d dummyLogger) Debug3(string, ...interface{}) {}
func (d dummyLogger) Debug2(string, ...interface{}) {}
func (d dummyLogger) Debug(...interface{})          {}
func (d dummyLogger) Info(...interface{})           {}
func (d dummyLogger) Warning(...interface{})        {}
func (d dummyLogger) Error(...interface{})          {}
func (d dummyLogger) Fatal(args ...interface{}) {
	fmt.Fprint(os.Stderr, args...)
	os.Exit(1)
}
func (d dummyLogger) Panic(args ...interface{}) {
	panic(fmt.Sprint(args...))
}
func (d dummyLogger) Debug3ln(...interface{})  {}
func (d dummyLogger) Debug2ln(...interface{})  {}
func (d dummyLogger) Debugln(...interface{})   {}
func (d dummyLogger) Infoln(...interface{})    {}
func (d dummyLogger) Warningln(...interface{}) {}
func (d dummyLogger) Errorln(...interface{})   {}
func (d dummyLogger) Fatalln(args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
	os.Exit(1)
}
func (d dummyLogger) Panicln(args ...interface{}) {
	panic(fmt.Sprintln(args...))
}
