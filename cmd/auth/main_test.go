package main

import (
	"os"
	"testing"

	"github.com/skip2/go-qrcode"
)

func TestParseOtpFromImage_HappyPath(t *testing.T) {
	otpUrl := "otpauth://totp/TestIssuer:tester@example.com?secret=JBSWY3DPEHPK3PXP&issuer=TestIssuer"
	tmpFile := "test_happy.png"
	defer os.Remove(tmpFile)

	err := qrcode.WriteFile(otpUrl, qrcode.Medium, 256, tmpFile)
	if err != nil {
		t.Fatalf("Failed to generate QR: %v", err)
	}

	u, err := parseOtpFromImage(tmpFile)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if u.Query().Get("secret") != "JBSWY3DPEHPK3PXP" {
		t.Errorf("Wrong secret: %s", u.Query().Get("secret"))
	}
}

func TestParseOtpFromImage_NonExistentFile(t *testing.T) {
	_, err := parseOtpFromImage("nonexistent.png")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

func TestParseOtpFromImage_InvalidFile(t *testing.T) {
	tmpFile := "test_invalid.txt"
	defer os.Remove(tmpFile)

	os.WriteFile(tmpFile, []byte("not an image"), 0644)
	_, err := parseOtpFromImage(tmpFile)
	if err == nil {
		t.Error("Expected error for invalid image file, got nil")
	}
}

func TestParseOtpFromImage_NoQRCode(t *testing.T) {
	tmpFile := "test_blank.png"
	defer os.Remove(tmpFile)

	// Create a blank QR image (not containing any QR code)
	err := qrcode.WriteFile("not-a-qr-code", qrcode.Low, 256, tmpFile)
	if err != nil {
		t.Fatalf("Failed to create image: %v", err)
	}

	// Manually corrupt the payload
	os.WriteFile(tmpFile, []byte("garbage content"), 0644)

	_, err = parseOtpFromImage(tmpFile)
	if err == nil {
		t.Error("Expected error for no QR code, got nil")
	}
}

func TestParseOtpFromImage_InvalidQRCodeScheme(t *testing.T) {
	tmpFile := "test_nonscheme.png"
	defer os.Remove(tmpFile)

	err := qrcode.WriteFile("https://example.com", qrcode.Medium, 256, tmpFile)
	if err != nil {
		t.Fatalf("Failed to create QR: %v", err)
	}

	_, err = parseOtpFromImage(tmpFile)
	if err == nil {
		t.Error("Expected error for non-otpauth QR code, got nil")
	}
}

func TestParseOtpFromImage_MissingSecret(t *testing.T) {
	tmpFile := "test_nosecret.png"
	defer os.Remove(tmpFile)

	otpURL := "otpauth://totp/TestIssuer:tester@example.com?issuer=TestIssuer"
	err := qrcode.WriteFile(otpURL, qrcode.Medium, 256, tmpFile)
	if err != nil {
		t.Fatalf("Failed to create QR: %v", err)
	}

	u, err := parseOtpFromImage(tmpFile)
	if err != nil {
		t.Fatalf("Unexpected error parsing QR code: %v", err)
	}

	secret := u.Query().Get("secret")
	if secret != "" {
		t.Errorf("Expected missing secret, got %s", secret)
	}
}
