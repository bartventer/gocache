package keymod

import (
	"testing"
)

func TestKey_String(t *testing.T) {
	k := Key("mykey")
	expected := "mykey"
	if k.String() != expected {
		t.Errorf("expected %s, got %s", expected, k.String())
	}
}

func TestKey_Prefix(t *testing.T) {
	k := Key("mykey")
	prefixed := k.Prefix("prefix_")
	expected := Key("prefix_mykey")
	if prefixed != expected {
		t.Errorf("expected %s, got %s", expected, prefixed)
	}
}

func TestKey_Suffix(t *testing.T) {
	k := Key("mykey")
	suffixed := k.Suffix("_suffix")
	expected := Key("mykey_suffix")
	if suffixed != expected {
		t.Errorf("expected %s, got %s", expected, suffixed)
	}
}

func TestKey_TagPrefix(t *testing.T) {
	k := Key("mykey")
	tagPrefixed := k.TagPrefix("tag")
	expected := Key("{tag}mykey")
	if tagPrefixed != expected {
		t.Errorf("expected %s, got %s", expected, tagPrefixed)
	}
}

func TestKey_TagSuffix(t *testing.T) {
	k := Key("mykey")
	tagSuffixed := k.TagSuffix("tag")
	expected := Key("mykey{tag}")
	if tagSuffixed != expected {
		t.Errorf("expected %s, got %s", expected, tagSuffixed)
	}
}
