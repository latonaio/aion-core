// Copyright (c) 2019-2020 Latona. All rights reserved.

package log

import (
	"fmt"
	"log"
	"os"

	"github.com/kelseyhightower/envconfig"
)

type Debug struct {
	Debug string `envconfig:"DEBUG" default:"false"`
}

// Logger is the interface for logging messages.
type Logger interface {
	Printf(format string, v ...interface{})
	Print(v ...interface{})
	Println(v ...interface{})
	Debugf(format string, v ...interface{})
	Debugln(v ...interface{})
	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})
}

// The set of default loggers for each log level.
// (singleton pattern)
var (
	defaultLogger = &logger{}
	debug         = GetEnv()
)

type logger struct {
	processName string
	pid         int
}

func GetEnv() *Debug {
	debug := Debug{}
	if err := envconfig.Process("", &debug); err != nil {
		log.Fatalf("Can not read debug option")
	}
	if debug.Debug == "true" {
		log.Printf("Debug mode on")
	}
	return &debug
}

func (l *logger) Printf(loglevel string, format string, v ...interface{}) {
	log.Printf(l.getFormat(loglevel)+format, v...)
}

func (l *logger) Print(loglevel string, v ...interface{}) {
	log.Print(append([]interface{}{l.getFormat(loglevel)}, v...)...)
}

func (l *logger) Println(loglevel string, v ...interface{}) {
	log.Println(append([]interface{}{l.getFormat(loglevel)}, v...)...)
}

func (l *logger) Fatalf(loglevel string, format string, v ...interface{}) {
	log.Fatalf(l.getFormat(loglevel)+format, v...)
}

func (l *logger) Fatal(loglevel string, v ...interface{}) {
	log.Fatal(append([]interface{}{l.getFormat(loglevel)}, v...)...)
}

func (l *logger) getFormat(loglevel string) string {
	// TODO: set log level
	return fmt.Sprintf("- %s - %5d - %s - ", loglevel, l.pid, l.processName)
}

func (l *logger) SetFormat(processName string) {
	l.processName = processName
	l.pid = os.Getegid()
	log.SetFlags(log.Ldate | log.Lmicroseconds)
}

// Printf writes a formatted message to the log.
func Printf(format string, v ...interface{}) {
	defaultLogger.Printf("INFO", format, v...)
}

// Print writes a message to the log.
func Print(v ...interface{}) {
	defaultLogger.Print("INFO", v...)
}

// Println writes a line to the log.
func Println(v ...interface{}) {
	defaultLogger.Println("INFO", v...)
}

func Debugf(format string, v ...interface{}) {
	if debug.Debug == "true" {
		defaultLogger.Printf("DEBUG", format, v...)
	}
}

func Debugln(v ...interface{}) {
	if debug.Debug == "true" {
		defaultLogger.Println("DEBUG", v...)
	}
}

// Fatal writes a message to the log and aborts.
func Fatal(v ...interface{}) {
	defaultLogger.Fatal("FATAL", v...)
}

// Fatalf writes a formatted message to the log and aborts.
func Fatalf(format string, v ...interface{}) {
	defaultLogger.Fatalf("FATAL", format, v...)
}

func SetFormat(processName string) {
	defaultLogger.SetFormat(processName)
}
