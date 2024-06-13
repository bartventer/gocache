package keymod

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashTagModifier(t *testing.T) {
	tests := []struct {
		name     string
		hashTag  string
		key      string
		expected string
	}{
		{"No brackets in key or hashTag", "tag", "key", "{tag}key"},
		{"Brackets in key", "tag", "{key}", "{tag}{key}"},
		{"Brackets in hashTag", "{tag}", "key", "key"},
		{"Brackets in key and hashTag", "{tag}", "{key}", "{key}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modifier := HashTagModifier(tt.hashTag)
			assert.Equal(t, tt.expected, modifier(tt.key))
		})
	}
}

func TestModifyKey(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		modifiers []KeyModifier
		expected  string
	}{
		{"No modifiers", "key", []KeyModifier{}, "key"},
		{"One modifier", "key", []KeyModifier{HashTagModifier("tag")}, "{tag}key"},
		{"Multiple modifiers", "key", []KeyModifier{HashTagModifier("tag1"), HashTagModifier("tag2")}, "{tag2}{tag1}key"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ModifyKey(tt.key, tt.modifiers...))
		})
	}
}
