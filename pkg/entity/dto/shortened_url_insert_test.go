package dto

import (
	"testing"
	"time"
)

func TestShortenedUrlInsert_Validate(t *testing.T) {
	futureTime := time.Now().Add(24 * time.Hour).Format("2006-01-02T15:04")
	pastTime := time.Now().Add(-24 * time.Hour).Format("2006-01-02T15:04")

	tests := []struct {
		name        string
		originalUrl string
		expiresAt   string
		customSlug  string
		description string
		password    string
		expectError bool
	}{
		{
			name:        "Valid input",
			originalUrl: "https://example.com/very/long/url",
			expiresAt:   futureTime,
			customSlug:  "",
			description: "",
			password:    "",
			expectError: false,
		},
		{
			name:        "Valid input with custom slug",
			originalUrl: "https://example.com",
			expiresAt:   futureTime,
			customSlug:  "my-promo",
			description: "",
			password:    "",
			expectError: false,
		},
		{
			name:        "Custom slug too long",
			originalUrl: "https://example.com",
			expiresAt:   futureTime,
			customSlug:  "this-is-too-long",
			description: "",
			password:    "",
			expectError: true,
		},
		{
			name:        "Custom slug invalid chars",
			originalUrl: "https://example.com",
			expiresAt:   futureTime,
			customSlug:  "my_promo!",
			description: "",
			password:    "",
			expectError: true,
		},
		{
			name:        "Empty URL",
			originalUrl: "",
			expiresAt:   futureTime,
			expectError: true,
		},
		{
			name:        "Invalid URL format",
			originalUrl: "not-a-url",
			expiresAt:   futureTime,
			expectError: true,
		},
		{
			name:        "Invalid URL scheme",
			originalUrl: "ftp://example.com",
			expiresAt:   futureTime,
			expectError: true,
		},
		{
			name:        "Empty expiration date",
			originalUrl: "https://example.com",
			expiresAt:   "",
			expectError: true,
		},
		{
			name:        "Invalid date format",
			originalUrl: "https://example.com",
			expiresAt:   "12-12-2025",
			expectError: true,
		},
		{
			name:        "Expiration date in the past",
			originalUrl: "https://example.com",
			expiresAt:   pastTime,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := NewShortenedUrlInsert(tt.originalUrl, tt.expiresAt, tt.customSlug, tt.description, tt.password)
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
