// Copyright 2016 The Upspin Authors. All rights reserved.
// Copyright 2017 The Tapr Authors. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions
// are met:
//
//    * Redistributions of source code must retain the above copyright
//      notice, this list of conditions and the following disclaimer.
//    * Redistributions in binary form must reproduce the above
//      copyright notice, this list of conditions and the following
//      disclaimer in the documentation and/or other materials provided
//      with the distribution.
//    * Neither the name of Google Inc. nor the names of its
//      contributors may be used to endorse or promote products derived
//      from this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

// Package log exports logging primitives that log to stderr.
package log // import "tapr.space/log"

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

// Logger is the interface for logging messages.
type Logger interface {
	// Printf writes a formated message to the log.
	Printf(format string, v ...interface{})

	// Print writes a message to the log.
	Print(v ...interface{})

	// Println writes a line to the log.
	Println(v ...interface{})

	// Fatal writes a message to the log and aborts.
	Fatal(v ...interface{})

	// Fatalf writes a formated message to the log and aborts.
	Fatalf(format string, v ...interface{})
}

// Level represents the level of logging.
type Level int

// Different levels of logging.
const (
	DebugLevel Level = iota
	InfoLevel
	WarningLevel
	ErrorLevel
	DisabledLevel
)

// ExternalLogger describes a service that processes logs.
type ExternalLogger interface {
	Log(Level, string)
	Flush()
}

// Event represents a log event.
type Event struct {
	Level   Level
	Message string
}

// The set of default loggers for each log level.
var (
	Debug   = &logger{DebugLevel}
	Info    = &logger{InfoLevel}
	Warning = &logger{WarningLevel}
	Error   = &logger{ErrorLevel}
)

var (
	currentLevel  = InfoLevel
	defaultLogger = newDefaultLogger(os.Stderr)
	external      ExternalLogger

	mu       sync.RWMutex
	monitors []ExternalLogger
)

func newDefaultLogger(w io.Writer) Logger {
	return log.New(w, "", log.Ltime)
}

// Register connects an ExternalLogger to the default logger. This may only be
// called once.
func Register(e ExternalLogger) {
	if external != nil {
		panic("cannot register second external logger")
	}

	external = e
}

// Monitor attaches an ExternalLogger to the default logger. Logged
// messages will be relayed to the list of attached external loggers.
func Monitor(e ExternalLogger) {
	mu.Lock()
	defer mu.Unlock()

	monitors = append(monitors, e)
}

// SetOutput sets the default loggers to write to w.
// If w is nil, the default loggers are disabled.
func SetOutput(w io.Writer) {
	if w == nil {
		defaultLogger = nil
	} else {
		defaultLogger = newDefaultLogger(w)
	}
}

type logger struct {
	level Level
}

var _ Logger = (*logger)(nil)

// Printf writes a formatted message to the log.
func (l *logger) Printf(format string, v ...interface{}) {
	if l.level < currentLevel {
		return // Don't log at lower levels.
	}

	if external != nil {
		external.Log(l.level, fmt.Sprintf(format, v...))
	}

	mu.RLock()
	for _, m := range monitors {
		go func(m ExternalLogger) {
			m.Log(l.level, fmt.Sprintf(format, v...))
		}(m)
	}
	mu.RUnlock()

	if defaultLogger != nil {
		defaultLogger.Printf(format, v...)
	}
}

// Print writes a message to the log.
func (l *logger) Print(v ...interface{}) {
	if l.level < currentLevel {
		return // Don't log at lower levels.
	}

	if external != nil {
		external.Log(l.level, fmt.Sprint(v...))
	}

	mu.RLock()
	for _, m := range monitors {
		go func(m ExternalLogger) {
			m.Log(l.level, fmt.Sprint(v...))
		}(m)
	}
	mu.RUnlock()

	if defaultLogger != nil {
		defaultLogger.Print(v...)
	}
}

// Println writes a line to the log.
func (l *logger) Println(v ...interface{}) {
	if l.level < currentLevel {
		return // Don't log at lower levels.
	}

	if external != nil {
		external.Log(l.level, fmt.Sprintln(v...))
	}

	mu.RLock()
	for _, m := range monitors {
		go func(m ExternalLogger) {
			m.Log(l.level, fmt.Sprintln(v...))
		}(m)
	}
	mu.RUnlock()

	if defaultLogger != nil {
		defaultLogger.Println(v...)
	}
}

// Fatal writes a message to the log and aborts, regardless of the current log level.
func (l *logger) Fatal(v ...interface{}) {
	if external != nil {
		external.Log(l.level, fmt.Sprint(v...))
		external.Flush()
	}

	mu.RLock()
	for _, m := range monitors {
		m.Log(l.level, fmt.Sprint(v...))
		m.Flush()
	}
	mu.RUnlock()

	if defaultLogger != nil {
		defaultLogger.Fatal(v...)
	} else {
		log.Fatal(v...)
	}
}

// Fatalf writes a formatted message to the log and aborts, regardless of the
// current log level.
func (l *logger) Fatalf(format string, v ...interface{}) {
	if external != nil {
		external.Log(l.level, fmt.Sprintf(format, v...))
		external.Flush()
	}

	mu.RLock()
	for _, m := range monitors {
		m.Log(l.level, fmt.Sprintf(format, v...))
		m.Flush()
	}
	mu.RUnlock()

	if defaultLogger != nil {
		defaultLogger.Fatalf(format, v...)
	} else {
		log.Fatalf(format, v...)
	}
}

// Flush implements ExternalLogger.
func (l *logger) Flush() {
	if external != nil {
		external.Flush()
	}

	mu.RLock()
	for _, m := range monitors {
		m.Flush()
	}
	mu.RUnlock()
}

// String returns the name of the logger.
func (l *logger) String() string {
	return toString(l.level)
}

func toString(level Level) string {
	switch level {
	case InfoLevel:
		return "info"
	case DebugLevel:
		return "debug"
	case WarningLevel:
		return "warning"
	case ErrorLevel:
		return "error"
	case DisabledLevel:
		return "disabled"
	}
	return "unknown"
}

func toLevel(level string) (Level, error) {
	switch level {
	case "info":
		return InfoLevel, nil
	case "debug":
		return DebugLevel, nil
	case "warning":
		return WarningLevel, nil
	case "error":
		return ErrorLevel, nil
	case "disabled":
		return DisabledLevel, nil
	}
	return DisabledLevel, fmt.Errorf("invalid log level %q", level)
}

// GetLevel returns the current logging level.
func GetLevel() string {
	return toString(currentLevel)
}

// SetLevel sets the current level of logging.
func SetLevel(level string) error {
	l, err := toLevel(level)
	if err != nil {
		return err
	}

	currentLevel = l

	return nil
}

// At returns whether the level will be logged currently.
func At(level string) bool {
	l, err := toLevel(level)
	if err != nil {
		return false
	}

	return currentLevel <= l
}

// Printf writes a formatted message to the log.
func Printf(format string, v ...interface{}) {
	Info.Printf(format, v...)
}

// Print writes a message to the log.
func Print(v ...interface{}) {
	Info.Print(v...)
}

// Println writes a line to the log.
func Println(v ...interface{}) {
	Info.Println(v...)
}

// Fatal writes a message to the log and aborts.
func Fatal(v ...interface{}) {
	Info.Fatal(v...)
}

// Fatalf writes a formatted message to the log and aborts.
func Fatalf(format string, v ...interface{}) {
	Info.Fatalf(format, v...)
}

// Flush flushes the external logger, if any.
func Flush() {
	if external != nil {
		external.Flush()
	}

	mu.RLock()
	for _, m := range monitors {
		m.Flush()
	}
	mu.RUnlock()
}
