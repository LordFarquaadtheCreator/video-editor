package logger

import (
	"fmt"
	"os"
)

var debugEnabled bool

// SetDebug enables or disables debug logging
func SetDebug(enabled bool) {
	debugEnabled = enabled
}

// Debug logs only if debug mode is enabled
func Debug(format string, args ...interface{}) {
	if debugEnabled {
		fmt.Fprintf(os.Stderr, "[DEBUG] "+format+"\n", args...)
	}
}

// Info logs informational messages
func Info(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[INFO] "+format+"\n", args...)
}

// Error logs error messages
func Error(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[ERROR] "+format+"\n", args...)
}

// Success logs success messages
func Success(format string, args ...interface{}) {
	fmt.Fprintf(os.Stdout, format+"\n", args...)
}
