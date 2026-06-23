package logs

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

// Log level
type LogLevel int

// Log level value
const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	DISABLED
)

const (
	reset   = "\033[0m"
	red     = "\033[31m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	blue    = "\033[34m"
	magenta = "\033[35m"
	cyan    = "\033[36m"
	white   = "\033[37m"
)

var start = time.Now()
var loglevel = DEBUG

// Set application log level
func SetLogLevel(level LogLevel) {
	loglevel = level
}

// Get application log level
func GetLogLevel() LogLevel {
	return loglevel
}

// Allias to Log(DEBUG, format, a...) with timestamp
func Debug(format string, a ...any) {
	Log(DEBUG, timeString()+format, a...)
}

// Allias to Log(INFO, format, a...)
func Info(format string, a ...any) {
	Log(INFO, format, a...)
}

// Allias to Log(WARN, format, a...)
func Warn(format string, a ...any) {
	Log(WARN, format, a...)
}

// Allias to Log(ERROR, format, a...)
func Error(format string, a ...any) {
	Log(ERROR, format, a...)
}

// Allias to Log(level, format, a...) with timestamp and caller name
func Trace(level LogLevel, format string, a ...any) {
	Log(level, timeString()+" "+callerName(2)+": "+format, a...)
}

// Allias to Log(level, format, a...) with timestamp and caller names
func TraceFrom(level LogLevel, format string, a ...any) {
	Log(level, timeString()+" "+callerName(3)+" > "+callerName(2)+": "+format, a...)
}

// Log message to console
func Log(level LogLevel, format string, a ...any) {
	if loglevel > level {
		return
	}
	var prefix string
	switch level {
	case DEBUG:
		prefix = magenta + "[DEBUG] " + reset
	case INFO:
		prefix = blue + "[INFO] " + reset
	case WARN:
		prefix = yellow + "[WARN] " + reset
	case ERROR:
		prefix = red + "[ERROR] " + reset
	}
	fmt.Printf(prefix+format+"\n", a...)
}

func callerName(skip int) string {
	pc, _, _, ok := runtime.Caller(skip)
	if !ok {
		return "unknown"
	}
	_, name, _ := strings.Cut(runtime.FuncForPC(pc).Name(), ".")
	return name
}

func timeString() string {
	return "(" + time.Since(start).String() + ") "
}
