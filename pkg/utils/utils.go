package utils

import (
	"fmt"
)

// FormatMessage returns a formatted string combining a message and data
func FormatMessage(message string, data interface{}) string {
    return fmt.Sprintf("%s: %v", message, data)
}
