package utils

import (
	"fmt"
	"log"
	"os"
)

// Logging that supports log levels and per package and per file
// logging control.

var logLevel = DebugLevel // Default log level

type level int

const (
	// DebugLevel logs a message at debug level
	DebugLevel level = iota
	// InfoLevel logs a message at info level
	InfoLevel
	// FatalLevel logs a message, then calls os.Exit(1).
	FatalLevel
)

func shouldLog(ll level) bool {
	if logLevel <= ll {
		return true
	}
	return false
}


type dfdLogger struct {
	li       *log.Logger
	ld       *log.Logger
	le       *log.Logger
}

func (l *dfdLogger) Info(args ...interface{}) {
	l.li.Output(3, fmt.Sprint(args...))
}
func (l *dfdLogger) Infoln(args ...interface{}) {
	l.li.Output(3, fmt.Sprintln(args...))
}
func (l *dfdLogger) Infof(format string, args ...interface{}) {
	l.li.Output(3, fmt.Sprintf(format, args...))
}

func (l *dfdLogger) Debug(args ...interface{}) {
	l.ld.Output(3, fmt.Sprint(args...))
}
func (l *dfdLogger) Debugln(args ...interface{}) {
	l.ld.Output(3, fmt.Sprintln(args...))
}
func (l *dfdLogger) Debugf(format string, args ...interface{}) {
	l.ld.Output(3, fmt.Sprintf(format, args...))
}

func (l *dfdLogger) Fatal(args ...interface{}) {
	l.le.Output(3, fmt.Sprint(args...))
}
func (l *dfdLogger) Fatalln(args ...interface{}) {
	l.le.Output(3, fmt.Sprintln(args...))
}
func (l *dfdLogger) Fatalf(format string, args ...interface{}) {
	l.le.Output(3, fmt.Sprintf(format, args...))
}

// Fatal will issue a log when level is <= FatalLevel
func Fatal(args ...interface{}) {
	if shouldLog(FatalLevel) {
		defaultLogger.Fatal(args...)
	}
	os.Exit(1)
}

// Fatalln will issue a log when level is <= FatalLevel
func Fatalln(args ...interface{}) {
	if shouldLog(FatalLevel) {
		defaultLogger.Fatalln(args...)
	}
	os.Exit(1)
}

// Fatalf will issue a formattted log when level is <= FatalLevel
func Fatalf(format string, args ...interface{}) {
	if shouldLog(FatalLevel) {
		defaultLogger.Fatalf(format, args...)
	}
	os.Exit(1)
}

func Debug(args ...interface{}) {
	if shouldLog(DebugLevel) {
		defaultLogger.Debug(args...)
	}
}

// Debugln will issue a log when level is <= DebugLevel
func Debugln(args ...interface{}) {
	if shouldLog(DebugLevel) {
		defaultLogger.Debugln(args...)
	}
}

// Debugf will issue a formatted log when level is <= DebugLevel
func Debugf(format string, args ...interface{}) {
	if shouldLog(DebugLevel) {
		defaultLogger.Debugf(format, args...)
	}
}

// Info will issue a log when level is <= InfoLevel
func Info(args ...interface{}) {
	if shouldLog(InfoLevel) {
		defaultLogger.Info(args...)
	}
}

// Infoln will issue a log when level is <= InfoLevel
func Infoln(args ...interface{}) {
	if shouldLog(InfoLevel) {
		defaultLogger.Infoln(args...)
	}
}

// Infof will issue a formattted log when level is <= InfoLevel
func Infof(format string, args ...interface{}) {
	if shouldLog(InfoLevel) {
		defaultLogger.Infof(format, args...)
	}
}


