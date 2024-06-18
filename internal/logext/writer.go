// Package logext provides a utility for logging debug messages.
package logext

import (
	"flag"
	"io"
	"log"
)

var debug = flag.Bool("gocache-debug", false, "enable debug logging")

// Logger is an interface for logging debug messages.
type Logger interface {
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}

// NewLogger returns a new logger.
// If the debug flag is set, it logs messages to the provided output.
// Otherwise, it discards all log messages.
func NewLogger(output io.Writer) Logger {
	if !*debug {
		output = io.Discard
	}
	logger := log.New(output, "[gocache] ", log.LstdFlags|log.Lshortfile|log.Lmicroseconds)
	return logger
}
