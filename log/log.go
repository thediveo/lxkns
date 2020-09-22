package log

import "fmt"

type Level uint32

// Different logging levels; please note that lxkns only uses a subset of them.
// By coincidence, these levels match the logging levels of logus ;)
const (
	PanicLevel Level = iota
	FatalLevel
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
	TraceLevel
)

func SetLevel(level Level) {
	loglevel = level
}

var loglevel = InfoLevel

//
var Loggerfn func(level Level, msg string)

func Errorf(format string, args ...interface{}) {
	logfifenabled(ErrorLevel, format, args...)
}

func Warnf(format string, args ...interface{}) {
	logfifenabled(WarnLevel, format, args...)
}

func Infof(format string, args ...interface{}) {
	logfifenabled(InfoLevel, format, args...)
}

func Debugf(format string, args ...interface{}) {
	logfifenabled(DebugLevel, format, args...)
}

func logfifenabled(level Level, format string, args ...interface{}) {
	if Loggerfn != nil && level <= loglevel {
		Loggerfn(level, fmt.Sprintf(format, args...))
	}
}
