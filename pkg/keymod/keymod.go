// Package keymod provides functions for modifying keys.
//
// This package is particularly useful when working with distributed caching systems like Redis Cluster,
// where keys containing "{hashTag}" are ensured to exist on the same node. This allows related keys to be
// stored together, enabling multi-key operations like transactions and Lua scripts.
package keymod

import (
	"strings"
)

// Key represents a key in a distributed caching system.
type Key string

func (k Key) String() string {
	return string(k)
}

// Prefix prepends the given text to the key.
func (k Key) Prefix(text string) Key {
	var sb strings.Builder
	sb.WriteString(text)
	sb.WriteString(string(k))
	return Key(sb.String())
}

// Suffix appends the given text to the key.
func (k Key) Suffix(text string) Key {
	var sb strings.Builder
	sb.WriteString(string(k))
	sb.WriteString(text)
	return Key(sb.String())
}

// TagPrefix wraps the given text in curly braces and prepends it to the key.
func (k Key) TagPrefix(text string) Key {
	return k.Prefix("{" + text + "}")
}

// TagSuffix wraps the given text in curly braces and appends it to the key.
func (k Key) TagSuffix(text string) Key {
	return k.Suffix("{" + text + "}")
}
