package keymod

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithHashTag(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		key      string
		expected string
	}{
		{"No brackets in key or hashTag", "tag", "key", "{tag}key"},
		{"Brackets in key", "tag", "{key}", "{tag}{key}"},
		{"Brackets in hashTag", "{tag}", "key", "{tag}key"},
		{"Brackets in key and hashTag", "{tag}", "{key}", "{tag}{key}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modifier := WithHashTag(tt.text)
			assert.Equal(t, tt.expected, modifier(tt.key))
		})
	}
}

func TestWithPrefix(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		key      string
		expected string
	}{
		{"No prefix", "", "key", "key"},
		{"With prefix", "prefix", "key", "prefix:key"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modifier := WithPrefix(tt.prefix, ":")
			assert.Equal(t, tt.expected, modifier(tt.key))
		})
	}
}

func TestWithSuffix(t *testing.T) {
	tests := []struct {
		name     string
		suffix   string
		key      string
		expected string
	}{
		{"No suffix", "", "key", "key"},
		{"With suffix", "suffix", "key", "key:suffix"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modifier := WithSuffix(tt.suffix, ":")
			assert.Equal(t, tt.expected, modifier(tt.key))
		})
	}
}

func TestWithChain(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		modifiers []Mod
		expected  string
	}{
		{"No modifiers", "key", []Mod{}, "key"},
		{"One modifier", "key", []Mod{WithHashTag("tag")}, "{tag}key"},
		{"Multiple modifiers", "key", []Mod{WithHashTag("tag1"), WithPrefix("prefix", ":")}, "prefix:{tag1}key"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modifier := WithChain(tt.modifiers...)
			assert.Equal(t, tt.expected, modifier(tt.key))
		})
	}
}

func TestModify(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		modifiers []Mod
		expected  string
	}{
		{"No modifiers", "key", []Mod{}, "key"},
		{"One modifier", "key", []Mod{WithHashTag("tag")}, "{tag}key"},
		{"Multiple modifiers", "key", []Mod{WithHashTag("tag1"), WithHashTag("tag2")}, "{tag2}{tag1}key"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, Modify(tt.key, tt.modifiers...))
		})
	}
}
