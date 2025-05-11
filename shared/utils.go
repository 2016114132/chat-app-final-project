// Shared helpers for formatting, validation, and command parsing used by both TCP and UDP clients/servers.
package shared

import (
	"fmt"
	"strings"
)

// SanitizeInput trims leading and trailing whitespace (spaces, tabs, newlines) from user input.
// This helps prevent accidental message formatting issues and removes empty lines.
func SanitizeInput(input string) string {
	return strings.TrimSpace(input)
}

// FormatMessage constructs a standardized chat message format that includes the sender's name.
// Example output: "[John]: Hello world"
func FormatMessage(sender, message string) string {
	sanitized := SanitizeInput(message)

	parts := strings.SplitN(sanitized, "|", 2)
	if len(parts) == 2 {
		sanitized = fmt.Sprintf("%s|[%s]: %s", parts[0], sender, parts[1])
	} else {
		sanitized = fmt.Sprintf("[%s]: %s", sender, sanitized)
	}

	return sanitized
}

// IsCommand checks if the input string starts with a forward slash (e.g., "/name") and matches the given command.
// It is case-insensitive.
// Used to detect commands like "/name", "/ping", etc.
func IsCommand(input, command string) bool {
	return strings.HasPrefix(strings.ToLower(input), "/"+command)
}

// FormatName sanitizes and validates a user's nickname.
// If the nickname is blank or only whitespace, it defaults to "Anonymous".
func FormatName(name string) string {
	n := SanitizeInput(name)
	if n == "" {
		return "Anonymous"
	}
	return n
}
