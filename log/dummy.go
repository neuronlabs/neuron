package log

import (
	"fmt"
	"os"
)

// Compile time check for dummyLogger to implement ExtendedLeveledLogger.
var _ ExtendedLeveledLogger = dummyLogger{}

type dummyLogger struct{}

func (d dummyLogger) Print(args ...interface{})                   {}
func (d dummyLogger) Printf(format string, args ...interface{})   {}
func (d dummyLogger) Println(args ...interface{})                 {}
func (d dummyLogger) Debug3f(format string, args ...interface{})  {}
func (d dummyLogger) Debug2f(format string, args ...interface{})  {}
func (d dummyLogger) Debugf(format string, args ...interface{})   {}
func (d dummyLogger) Infof(format string, args ...interface{})    {}
func (d dummyLogger) Warningf(format string, args ...interface{}) {}
func (d dummyLogger) Errorf(format string, args ...interface{})   {}
func (d dummyLogger) Fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}
func (d dummyLogger) Panicf(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args...))
}
func (d dummyLogger) Debug3(format string, args ...interface{}) {}
func (d dummyLogger) Debug2(format string, args ...interface{}) {}
func (d dummyLogger) Debug(args ...interface{})                 {}
func (d dummyLogger) Info(args ...interface{})                  {}
func (d dummyLogger) Warning(args ...interface{})               {}
func (d dummyLogger) Error(args ...interface{})                 {}
func (d dummyLogger) Fatal(args ...interface{}) {
	fmt.Fprint(os.Stderr, args...)
	os.Exit(1)
}
func (d dummyLogger) Panic(args ...interface{}) {
	panic(fmt.Sprint(args...))
}
func (d dummyLogger) Debug3ln(args ...interface{})  {}
func (d dummyLogger) Debug2ln(args ...interface{})  {}
func (d dummyLogger) Debugln(args ...interface{})   {}
func (d dummyLogger) Infoln(args ...interface{})    {}
func (d dummyLogger) Warningln(args ...interface{}) {}
func (d dummyLogger) Errorln(args ...interface{})   {}
func (d dummyLogger) Fatalln(args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
	os.Exit(1)
}
func (d dummyLogger) Panicln(args ...interface{}) {
	panic(fmt.Sprintln(args...))
}
