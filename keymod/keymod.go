// Package keymod provides functions for modifying keys.
//
// This package is particularly useful when working with distributed caching systems like Redis Cluster,
// where keys containing "{hashTag}" are ensured to exist on the same node. This allows related keys to be
// stored together, enabling multi-key operations like transactions and Lua scripts.
package keymod

import "strings"

// KeyModifier is a function that modifies a key.
//
// This can be used to add prefixes, suffixes, or hash tags to keys.
type KeyModifier func(string) string

// WithHashTag wraps the given text in curly braces and prepends it to the key.
//
// This is useful when working with Redis Cluster, as keys with the same hash tag
// are guaranteed to be on the same node. This allows for multi-key operations to
// be performed atomically.
//
// Example:
//
//	userHashTagModifier := WithHashTag("user123")
//	userKey := userHashTagModifier("profile") // "{user123}profile"
func WithHashTag(text string) KeyModifier {
	return func(key string) string {
		return "{" + strings.Trim(text, "{}") + "}" + key
	}
}

// WithPrefix prepends the given prefix to the key.
//
// This is useful when working with namespaced keys, where the prefix represents a namespace.
// The separator is used to separate the prefix from the key.
//
// Example:
//
//	userPrefixModifier := WithPrefix("user123", ":")
//	userKey := userPrefixModifier("profile") // "user123:profile"
func WithPrefix(prefix string, separator string) KeyModifier {
	return func(key string) string {
		if prefix == "" {
			return key
		}
		return prefix + separator + key
	}
}

// WithSuffix appends the given suffix to the key.
//
// This is useful when working with namespaced keys, where the suffix represents a namespace.
// The separator is used to separate the key from the suffix.
//
// Example:
//
//	userSuffixModifier := WithSuffix("profile", ":")
//	userKey := userSuffixModifier("user123") // "user123:profile"
func WithSuffix(suffix string, separator string) KeyModifier {
	return func(key string) string {
		if suffix == "" {
			return key
		}
		return key + separator + suffix
	}
}

// WithChain chains multiple KeyModifier functions together into a single KeyModifier.
//
// This can be used to create complex key modifications in a more readable way.
//
// Example:
//
//	modifier := WithChain(PrefixModifier("user123", ":"), HashTagModifier("user123"))
//	key := modifier("profile") // "{user123}user123:profile"
func WithChain(modifiers ...KeyModifier) KeyModifier {
	return func(key string) string {
		for _, modifier := range modifiers {
			key = modifier(key)
		}
		return key
	}
}

// Modify applies the given KeyModifier functions to a key.
//
// This can be used to apply multiple modifications to a key, such as adding a prefix and a hash tag.
//
// Example:
//
//	key := Modify("profile", PrefixModifier("user123", ":"), HashTagModifier("user123")) // "{user123}user123:profile"
func Modify(key string, modifiers ...KeyModifier) string {
	for _, modifier := range modifiers {
		key = modifier(key)
	}
	return key
}
