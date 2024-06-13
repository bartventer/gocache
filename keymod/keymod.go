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

// HashTagModifier returns a KeyModifier that adds a hash tag to a key.
//
// This is useful when working with Redis Cluster, as keys with the same hash tag are guaranteed to be
// on the same node. This allows for multi-key operations to be performed atomically.
//
// Example:
//
//	userHashTagModifier := HashTagModifier("user123")
//	userKey := userHashTagModifier("profile") // "{user123}profile"
func HashTagModifier(hashTag string) KeyModifier {
	return func(key string) string {
		if strings.Contains(hashTag, "{") || strings.Contains(hashTag, "}") {
			return key
		}
		return "{" + hashTag + "}" + key
	}
}

// ModifyKey applies the given KeyModifier functions to a key.
//
// This can be used to apply multiple modifications to a key, such as adding a prefix and a hash tag.
func ModifyKey(key string, modifiers ...KeyModifier) string {
	for _, modifier := range modifiers {
		key = modifier(key)
	}
	return key
}
