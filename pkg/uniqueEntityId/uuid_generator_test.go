package uniqueEntityId

import (
	"testing"
)

func TestNewID(t *testing.T) {
	id := NewID()

	if id.String() == "" {
		t.Error("Expected a non-empty UUID string")
	}

	// Should generate different IDs
	id2 := NewID()
	if id == id2 {
		t.Error("Expected two generated IDs to be different")
	}
}

func TestParseID(t *testing.T) {
	validIdStr := NewID().String()

	parsed, err := ParseID(validIdStr)
	if err != nil {
		t.Errorf("Expected no error when parsing valid UUID, got: %v", err)
	}

	if parsed.String() != validIdStr {
		t.Errorf("Expected parsed ID to match original: got %s, want %s", parsed.String(), validIdStr)
	}

	_, err = ParseID("invalid-uuid")
	if err == nil {
		t.Error("Expected error when parsing invalid UUID, got nil")
	}
}
