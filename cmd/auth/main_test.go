package main

import (
	"os"
	"testing"

	"github.com/skip2/go-qrcode"
)

func TestParseOtpFromImage_Table(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		expectError  bool
		expectSecret string
	}{
		{
			name:         "happy path",
			url:          "otpauth://totp/TestIssuer:tester@example.com?secret=JBSWY3DPEHPK3PXP&issuer=TestIssuer",
			expectError:  false,
			expectSecret: "JBSWY3DPEHPK3PXP",
		},
		{
			name:         "missing secret",
			url:          "otpauth://totp/TestIssuer:tester@example.com?issuer=TestIssuer",
			expectError:  false, // still parsable
			expectSecret: "",
		},
		{
			name:        "not otpauth scheme",
			url:         "https://example.com",
			expectError: true,
		},
		{
			name:        "invalid image content",
			url:         "", // will corrupt image file later
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := "test_" + tt.name + ".png"
			defer os.Remove(tmpFile)

			if tt.name == "invalid image content" {
				// Write garbage to simulate corrupt image
				_ = os.WriteFile(tmpFile, []byte("this is not a real image"), 0644)
			} else {
				err := qrcode.WriteFile(tt.url, qrcode.Medium, 256, tmpFile)
				if err != nil {
					t.Fatalf("Failed to write QR code: %v", err)
				}
			}

			u, err := parseOtpFromImage(tmpFile)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Optional check for secret value
			if got := u.Query().Get("secret"); got != tt.expectSecret {
				t.Errorf("Expected secret %q, got %q", tt.expectSecret, got)
			}
		})
	}
}
