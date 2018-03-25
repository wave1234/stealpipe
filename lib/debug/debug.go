package debug

import (
	"fmt"
)

var LineDebug = true // print the line num of source code

const (
	OFF   = 0
	FATAL = 100
	ERROR = 200
	WARN  = 300
	INFO  = 400
	DEBUG = 500
	TRACE = 600
	ALL   = 9999
)

var DebugLevel = DEBUG

func file_line() string {
	if LineDebug == false {
		return ""
	}
	return ""
}

func LogLevel(l int) {
	DebugLevel = l
}

func Fatal(a ...interface{}) {
	if DebugLevel >= FATAL {
		fmt.Println("[FATAL] "+file_line(), a)
	}
}

func Error(a ...interface{}) {
	if DebugLevel >= ERROR {
		fmt.Println("[ERROR] "+file_line(), a)
	}
}

func Warn(a ...interface{}) {
	if DebugLevel >= WARN {
		fmt.Println("[WARN] "+file_line(), a)
	}
}

func Info(a ...interface{}) {
	return
	if DebugLevel >= INFO {
		fmt.Println("[INFO] "+file_line(), a)
	}
}

func Debug(a ...interface{}) {
	return
	if DebugLevel >= DEBUG {
		fmt.Println("[DEBUG] "+file_line(), a)
	}
}

func Trace(a ...interface{}) {
	return
	if DebugLevel >= TRACE {
		fmt.Println("[TRACE] "+file_line(), a)
	}
}
