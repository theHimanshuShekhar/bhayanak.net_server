package logger

import (
	"log"
)

// Info logs informational messages
func Info(message string) {
    log.Printf("[INFO] %s\n", message)
}

// Error logs error messages
func Error(message string, err error) {
    log.Printf("[ERROR] %s: %v\n", message, err)
}
