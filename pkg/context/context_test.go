package context

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSetAndGetAuthUser(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	userID := "12345"

	// Ensure no user exists initially
	_, ok := GetAuthUser(req)
	if ok {
		t.Error("Expected no user in context initially")
	}

	// Set user
	req = SetAuthUser(req, userID)

	// Get user back
	gotID, ok := GetAuthUser(req)
	if !ok {
		t.Error("Expected to find user in context")
	}
	if gotID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, gotID)
	}
}

func TestIsCanceledError(t *testing.T) {
	if !IsCanceledError(context.Canceled) {
		t.Error("Expected true for context.Canceled")
	}

	if IsCanceledError(http.ErrServerClosed) {
		t.Error("Expected false for non-canceled error")
	}
}
