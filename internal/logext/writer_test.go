package logext

import (
	"bytes"
	"testing"
)

func TestNewLogger(t *testing.T) {
	t.Run("debug", func(t *testing.T) {
		var b bytes.Buffer
		*debug = true
		logger := NewLogger(&b)
		logger.Println("test message")
		if b.String() == "" {
			t.Error("Expected logger to write output")
		}
	})

	t.Run("no debug", func(t *testing.T) {
		var b bytes.Buffer
		*debug = false
		logger := NewLogger(&b)
		logger.Println("test message")
		if b.String() != "" {
			t.Error("Expected logger to not write output")
		}
	})
}
