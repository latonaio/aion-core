// Copyright (c) 2019-2020 Latona. All rights reserved.

package log

import (
	"fmt"
	"log"
	"os"
)

// Logger is the interface for logging messages.
type Logger interface {
	Printf(format string, v ...interface{})
	Print(v ...interface{})
	Println(v ...interface{})
	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})
}

// The set of default loggers for each log level.
// (singleton pattern)
var (
	defaultLogger = &logger{}
)

type logger struct {
	processName string
	pid         int
}

func (l *logger) Printf(format string, v ...interface{}) {
	log.Printf(l.getFormat()+format, v...)
}

func (l *logger) Print(v ...interface{}) {
	log.Print(append([]interface{}{l.getFormat()}, v...)...)
}

func (l *logger) Println(v ...interface{}) {
	log.Println(append([]interface{}{l.getFormat()}, v...)...)
}

func (l *logger) Fatalf(format string, v ...interface{}) {
	log.Fatalf(l.getFormat()+format, v...)
}

func (l *logger) Fatal(v ...interface{}) {
	log.Fatal(append([]interface{}{l.getFormat()}, v...)...)
}

func (l *logger) getFormat() string {
	// TODO: set log level
	return fmt.Sprintf("- DEBUG - %5d - %s - ", l.pid, l.processName)
}

func (l *logger) SetFormat(processName string) {
	l.processName = processName
	l.pid = os.Getegid()
	log.SetFlags(log.Ldate | log.Lmicroseconds)
}

// Printf writes a formatted message to the log.
func Printf(format string, v ...interface{}) {
	defaultLogger.Printf(format, v...)
}

// Print writes a message to the log.
func Print(v ...interface{}) {
	defaultLogger.Print(v...)
}

// Println writes a line to the log.
func Println(v ...interface{}) {
	defaultLogger.Println(v...)
}

// Fatal writes a message to the log and aborts.
func Fatal(v ...interface{}) {
	defaultLogger.Fatal(v...)
}

// Fatalf writes a formatted message to the log and aborts.
func Fatalf(format string, v ...interface{}) {
	defaultLogger.Fatalf(format, v...)
}

func SetFormat(processName string) {
	defaultLogger.SetFormat(processName)
}
