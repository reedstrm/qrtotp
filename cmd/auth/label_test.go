package main

import (
	"net/url"
	"testing"
)

func TestExtractIssuerAndLabel(t *testing.T) {
	tests := []struct {
		name         string
		rawURL       string
		expectIssuer string
		expectLabel  string
	}{
		{
			name:         "with issuer param and colon",
			rawURL:       "otpauth://totp/Example:alice@example.com?secret=XYZ&issuer=Example",
			expectIssuer: "Example",
			expectLabel:  "alice@example.com",
		},
		{
			name:         "with issuer param only",
			rawURL:       "otpauth://totp/alice@example.com?secret=XYZ&issuer=OnlyIssuer",
			expectIssuer: "OnlyIssuer",
			expectLabel:  "alice@example.com",
		},
		{
			name:         "colon-only fallback",
			rawURL:       "otpauth://totp/FallbackCorp:bob?secret=123",
			expectIssuer: "FallbackCorp",
			expectLabel:  "bob",
		},
		{
			name:         "no colon or issuer",
			rawURL:       "otpauth://totp/bob?secret=123",
			expectIssuer: "bob",
			expectLabel:  "bob",
		},
		{
			name:         "url-encoded",
			rawURL:       "otpauth://totp/Google%3Aross%40reedstrom.org?secret=ABC&issuer=Google",
			expectIssuer: "Google",
			expectLabel:  "ross@reedstrom.org",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := url.Parse(tt.rawURL)
			if err != nil {
				t.Fatalf("Invalid URL: %v", err)
			}

			gotIssuer, gotLabel := extractIssuerAndLabel(u)

			if gotIssuer != tt.expectIssuer {
				t.Errorf("Expected issuer %q, got %q", tt.expectIssuer, gotIssuer)
			}
			if gotLabel != tt.expectLabel {
				t.Errorf("Expected label %q, got %q", tt.expectLabel, gotLabel)
			}
		})
	}
}
