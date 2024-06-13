// Package testutil provides utilities for testing.
package testutil

import (
	"fmt"
	"testing"
	"time"
)

// UniqueKey returns a unique key for the test.
func UniqueKey(t *testing.T) string {
	t.Helper()
	return fmt.Sprintf("%s-%d", t.Name(), time.Now().UnixNano())
}
