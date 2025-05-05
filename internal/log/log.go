// internal/log/log.go
package log

import (
	"fmt"
	"os"
)

// Debugf prints a formatted debug message to stderr if debugging is enabled.
// It automatically adds a "DEBUG: " prefix.
// Note: This basic version doesn't check a global flag; the caller should.
func Debugf(format string, args ...interface{}) {
	// Prepend DEBUG prefix
	debugFormat := "DEBUG: " + format
	// Ensure newline if not present
	if len(debugFormat) > 0 && debugFormat[len(debugFormat)-1] != '\n' {
		debugFormat += "\n"
	}
	// Print to Stderr to separate from normal output
	fmt.Fprintf(os.Stderr, debugFormat, args...)
}

// TODO: Consider adding log levels (Info, Warn, Error) and configuration later.
