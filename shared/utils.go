// Shared helpers if needed. Maybe like formatting, logging, etc.package shared
package shared

import (
	"fmt"
	"strings"
)

// Trims and normalizes user input
func SanitizeInput(input string) string {
	return strings.TrimSpace(input)
}

// Formats a chat message with a name
func FormatMessage(sender, message string) string {
	sanitized := SanitizeInput(message)
	return fmt.Sprintf("[%s]: %s", sender, sanitized)
}

// Detect special commands like /name
func IsCommand(input, command string) bool {
	return strings.HasPrefix(strings.ToLower(input), "/"+command)
}

// Handle blank of bad name values
func FormatName(name string) string {
	n := SanitizeInput(name)
	if n == "" {
		return "Anonymous"
	}
	return n
}
