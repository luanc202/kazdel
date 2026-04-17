package dto

import (
	"testing"
)

func TestLoginRequest_Validate(t *testing.T) {
	tests := []struct {
		name        string
		username    string
		password    string
		expectError bool
	}{
		{
			name:        "Valid login request",
			username:    "validuser",
			password:    "ValidPass123!",
			expectError: false,
		},
		{
			name:        "Empty username",
			username:    "",
			password:    "ValidPass123!",
			expectError: true,
		},
		{
			name:        "Empty password",
			username:    "validuser",
			password:    "",
			expectError: true,
		},
		{
			name:        "Username too short",
			username:    "abc",
			password:    "ValidPass123!",
			expectError: true,
		},
		{
			name:        "Username too long",
			username:    "thisusernameiswaytoolongtoaccept",
			password:    "ValidPass123!",
			expectError: true,
		},
		{
			name:        "Invalid characters in username",
			username:    "user@name",
			password:    "ValidPass123!",
			expectError: true,
		},
		{
			name:        "Password too short",
			username:    "validuser",
			password:    "short",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := NewLoginRequest(tt.username, tt.password)
			err := req.Validate()

			if tt.expectError && err == nil {
				t.Errorf("Expected an error but got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got %v", err)
			}
		})
	}
}
