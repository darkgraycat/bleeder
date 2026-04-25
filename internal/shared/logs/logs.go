package logs

import "fmt"

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

func Info(format string, a ...any) {
	log(blue, "[INFO] ", format, a)
}

func Warn(format string, a ...any) {
	log(yellow, "[WARN] ", format, a)
}

func Error(format string, a ...any) {
	log(red, "[ERROR] ", format, a)
}

func Debug(format string, a ...any) {
	log(magenta, "[DEBUG] ", format, a)
}

func log(clr, prefix, format string, a ...any) {
	fmt.Printf(clr+prefix+reset+format+"\n", a...)
}
