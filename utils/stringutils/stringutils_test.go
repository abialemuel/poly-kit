package stringutils

import (
	"testing"
)

func TestToUpper(t *testing.T) {
	input := "hello"
	expected := "HELLO"
	result := ToUpper(input)
	if result != expected {
		t.Errorf("ToUpper(%q) = %q; want %q", input, result, expected)
	}
}

func TestToLower(t *testing.T) {
	input := "WORLD"
	expected := "world"
	result := ToLower(input)
	if result != expected {
		t.Errorf("ToLower(%q) = %q; want %q", input, result, expected)
	}
}

func TestReverse(t *testing.T) {
	input := "GoLang"
	expected := "gnaLoG"
	result := Reverse(input)
	if result != expected {
		t.Errorf("Reverse(%q) = %q; want %q", input, result, expected)
	}
}
