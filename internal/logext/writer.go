// Package logext provides a custom [log.Logger] interface for debug logging.
//
// Logging is controlled by the GOCACHE_DEBUG environment variable, set to
// "true" to enable debug logging.
package logext

import (
	"io"
	"log"
	"os"
)

// DebugEnvVar is the name of the environment variable that controls debug logging.
const DebugEnvVar = "GOCACHE_DEBUG"

// NewLogger returns a new logger.
// If the GOCACHE_DEBUG environment variable is set, it logs messages to the provided output.
// Otherwise, it discards all log messages.
func NewLogger(output io.Writer) *log.Logger {
	if os.Getenv(DebugEnvVar) != "true" {
		output = io.Discard
	}
	logger := log.New(output, "[gocache] ", log.LstdFlags|log.Lshortfile|log.Lmicroseconds)
	return logger
}
