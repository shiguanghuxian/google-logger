/*
Copyright 2016 Google Inc. All Rights Reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package logger offers simple cross platform logging for Windows and Linux.
// Available logging endpoints are event log (Windows), syslog (Linux), and
// an io.Writer.
package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

var (
	defaultLogger *Logger
	logLock       sync.Mutex
)

const (
	flags    = log.Ldate | log.Lmicroseconds | log.Lshortfile
	initText = "ERROR: Logging before logger.Init.\n"
)

func initialize() {
	defaultLogger = &Logger{
		infoLog:  log.New(os.Stderr, initText+"INFO: ", flags),
		warnLog:  log.New(os.Stderr, initText+"Warn: ", flags),
		errorLog: log.New(os.Stderr, initText+"ERROR: ", flags),
		fatalLog: log.New(os.Stderr, initText+"FATAL: ", flags),
	}
}

func init() {
	initialize()
}

// Init sets up logging and should be called before log functions, usually in
// the caller's main(). Default log functions can be called before Init(), but log
// output will only go to stderr (along with a warning).
// The first call to Init populates the default logger and returns the
// generated logger, subsequent calls to Init will only return the generated
// logger.
func Init(name string, verbose, systemLog bool, logFile io.Writer) *Logger {
	var il, el io.Writer
	if systemLog {
		var err error
		il, el, err = setup(name)
		if err != nil {
			log.Fatal(err)
		}
	}

	iLogs := []io.Writer{logFile}
	eLogs := []io.Writer{logFile, os.Stderr}
	if verbose {
		iLogs = append(iLogs, os.Stdout)
	}
	if il != nil {
		iLogs = append(iLogs, il)
	}
	if el != nil {
		eLogs = append(eLogs, el)
	}

	var l Logger
	l.infoLog = log.New(io.MultiWriter(iLogs...), "INFO: ", flags)
	l.warnLog = log.New(io.MultiWriter(eLogs...), "Warn: ", flags)
	l.errorLog = log.New(io.MultiWriter(eLogs...), "ERROR: ", flags)
	l.fatalLog = log.New(io.MultiWriter(eLogs...), "FATAL: ", flags)
	l.initialized = true

	logLock.Lock()
	defer logLock.Unlock()
	if !defaultLogger.initialized {
		defaultLogger = &l
	}

	return &l
}

type severity int

const (
	sInfo = iota
	sWarn
	sError
	sFatal
)

// A Logger represents an active logging object. Multiple loggers can be used
// simultaneously even if they are using the same same writers.
type Logger struct {
	infoLog     *log.Logger
	warnLog     *log.Logger
	errorLog    *log.Logger
	fatalLog    *log.Logger
	initialized bool
}

func (l *Logger) output(s severity, txt string) {
	logLock.Lock()
	defer logLock.Unlock()
	switch s {
	case sInfo:
		l.infoLog.Output(4, txt)
	case sWarn:
		l.warnLog.Output(4, txt)
	case sError:
		l.errorLog.Output(4, txt)
	case sFatal:
		l.fatalLog.Output(4, txt)
	default:
		panic(fmt.Sprintln("unrecognized severity:", s))
	}
}

// Info logs with the INFO severity.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Info(v ...interface{}) {
	l.output(sInfo, fmt.Sprint(v...))
}

// Infoln logs with the INFO severity.
// Arguments are handled in the manner of fmt.Println.
func (l *Logger) Infoln(v ...interface{}) {
	l.output(sInfo, fmt.Sprintln(v...))
}

// Infof logs with the INFO severity.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Infof(format string, v ...interface{}) {
	l.output(sInfo, fmt.Sprintf(format, v...))
}

// Warn logs with the Warn severity.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Warn(v ...interface{}) {
	l.output(sWarn, fmt.Sprint(v...))
}

// Warnln logs with the Warn severity.
// Arguments are handled in the manner of fmt.Println.
func (l *Logger) Warnln(v ...interface{}) {
	l.output(sWarn, fmt.Sprintln(v...))
}

// Warnf logs with the Warn severity.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Warnf(format string, v ...interface{}) {
	l.output(sWarn, fmt.Sprintf(format, v...))
}

// Error logs with the ERROR severity.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Error(v ...interface{}) {
	l.output(sError, fmt.Sprint(v...))
}

// Errorln logs with the ERROR severity.
// Arguments are handled in the manner of fmt.Println.
func (l *Logger) Errorln(v ...interface{}) {
	l.output(sError, fmt.Sprintln(v...))
}

// Errorf logs with the Error severity.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.output(sError, fmt.Sprintf(format, v...))
}

// Fatal logs with the Fatal severity, and ends with os.Exit(1).
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Fatal(v ...interface{}) {
	l.output(sFatal, fmt.Sprint(v...))
	os.Exit(1)
}

// Fatalln logs with the Fatal severity, and ends with os.Exit(1).
// Arguments are handled in the manner of fmt.Println.
func (l *Logger) Fatalln(v ...interface{}) {
	l.output(sFatal, fmt.Sprintln(v...))
	os.Exit(1)
}

// Fatalf logs with the Fatal severity, and ends with os.Exit(1).
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.output(sFatal, fmt.Sprintf(format, v...))
	os.Exit(1)
}

// Info uses the default logger and logs with the Info severity.
// Arguments are handled in the manner of fmt.Print.
func Info(v ...interface{}) {
	defaultLogger.output(sInfo, fmt.Sprint(v...))
}

// Infoln uses the default logger and logs with the Info severity.
// Arguments are handled in the manner of fmt.Println.
func Infoln(v ...interface{}) {
	defaultLogger.output(sInfo, fmt.Sprintln(v...))
}

// Infof uses the default logger and logs with the Info severity.
// Arguments are handled in the manner of fmt.Printf.
func Infof(format string, v ...interface{}) {
	defaultLogger.output(sInfo, fmt.Sprintf(format, v...))
}

// Warn uses the default logger and logs with the Warn severity.
// Arguments are handled in the manner of fmt.Print.
func Warn(v ...interface{}) {
	defaultLogger.output(sWarn, fmt.Sprint(v...))
}

// Warnln uses the default logger and logs with the Warn severity.
// Arguments are handled in the manner of fmt.Println.
func Warnln(v ...interface{}) {
	defaultLogger.output(sWarn, fmt.Sprintln(v...))
}

// Warnf uses the default logger and logs with the Warn severity.
// Arguments are handled in the manner of fmt.Printf.
func Warnf(format string, v ...interface{}) {
	defaultLogger.output(sWarn, fmt.Sprintf(format, v...))
}

// Error uses the default logger and logs with the Error severity.
// Arguments are handled in the manner of fmt.Print.
func Error(v ...interface{}) {
	defaultLogger.output(sError, fmt.Sprint(v...))
}

// Errorln uses the default logger and logs with the Error severity.
// Arguments are handled in the manner of fmt.Println.
func Errorln(v ...interface{}) {
	defaultLogger.output(sError, fmt.Sprintln(v...))
}

// Errorf uses the default logger and logs with the Error severity.
// Arguments are handled in the manner of fmt.Printf.
func Errorf(format string, v ...interface{}) {
	defaultLogger.output(sError, fmt.Sprintf(format, v...))
}

// Fatalln uses the default logger, logs with the Fatal severity,
// and ends with os.Exit(1).
// Arguments are handled in the manner of fmt.Print.
func Fatal(v ...interface{}) {
	defaultLogger.output(sFatal, fmt.Sprint(v...))
	os.Exit(1)
}

// Fatalln uses the default logger, logs with the Fatal severity,
// and ends with os.Exit(1).
// Arguments are handled in the manner of fmt.Println.
func Fatalln(v ...interface{}) {
	defaultLogger.output(sFatal, fmt.Sprintln(v...))
	os.Exit(1)
}

// Fatalf uses the default logger, logs with the Fatal severity,
// and ends with os.Exit(1).
// Arguments are handled in the manner of fmt.Printf.
func Fatalf(format string, v ...interface{}) {
	defaultLogger.output(sFatal, fmt.Sprintf(format, v...))
	os.Exit(1)
}
