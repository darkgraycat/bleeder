package logs

import (
	"fmt"
	"time"
)

var start = time.Now()
var loglevel = 0

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

func SetLogLevel(level int) {
	loglevel = level
}

func Error(format string, a ...any) {
	if loglevel > 3 {
		return
	}
	log(red, "[ERROR] ", format, a...)
}

func Warn(format string, a ...any) {
	if loglevel > 2 {
		return
	}
	log(yellow, "[WARN] ", format, a...)
}

func Info(format string, a ...any) {
	if loglevel > 1 {
		return
	}
	log(blue, "[INFO] ", format, a...)
}

func Debug(format string, a ...any) {
	if loglevel > 0 {
		return
	}
	timestr := "(" + time.Since(start).String() + ") "
	log(magenta, "[DEBUG] ", timestr+format, a...)
}

func log(clr, prefix, format string, a ...any) {
	fmt.Printf(clr+prefix+reset+format+"\n", a...)
}
