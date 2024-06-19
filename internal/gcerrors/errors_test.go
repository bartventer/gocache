package gcerrors

import (
	"errors"
	"testing"
)

func TestNew(t *testing.T) {
	err := errors.New("test error")
	gcErr := New(err)
	if gcErr.Error() != "gocache: test error" {
		t.Errorf("Expected 'gocache: test error', got '%s'", gcErr.Error())
	}
}

func TestNewWithScheme(t *testing.T) {
	err := errors.New("test error")
	gcErr := NewWithScheme("myscheme", err)
	if gcErr.Error() != "gocache/myscheme: test error" {
		t.Errorf("Expected 'gocache/myscheme: test error', got '%s'", gcErr.Error())
	}
}

func TestUnwrap(t *testing.T) {
	err := errors.New("test error")
	gcErr := New(err)
	unwrappedErr := gcErr.Unwrap()
	if unwrappedErr != err {
		t.Errorf("Expected unwrapped error to be '%v', got '%v'", err, unwrappedErr)
	}
}
