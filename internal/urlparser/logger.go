package urlparser

import "log"

// Logger is an interface for logging debug messages.
type Logger interface {
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}

// nopLogger is a no-op logger.
type nopLogger struct{}

func (nopLogger) Printf(format string, v ...interface{}) {}
func (nopLogger) Println(v ...interface{})               {}

func defaultLogger() Logger {
	logger := log.Default()
	logger.SetPrefix("[gocache] ")
	logger.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
	return logger
}
