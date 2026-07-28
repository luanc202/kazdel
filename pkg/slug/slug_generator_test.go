package slug

import (
	"strings"
	"testing"
)

func TestGenerateSlug(t *testing.T) {
	length := 8
	slug := GenerateSlug(length)

	if len(slug) != length {
		t.Errorf("Expected slug length %d, got %d", length, len(slug))
	}

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for _, char := range slug {
		if !strings.ContainsRune(charset, char) {
			t.Errorf("Slug contains invalid character: %c", char)
		}
	}

	// Test randomness by generating another one and ensuring it's different
	// (Flaky by nature if length is extremely small, but extremely unlikely to fail for length 8)
	slug2 := GenerateSlug(length)
	if slug == slug2 {
		t.Errorf("Generated two identical slugs consecutively: %s", slug)
	}
}
