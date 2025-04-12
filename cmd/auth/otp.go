package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/liyue201/goqr"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// processOtp wraps the parsing and extracting logic into a single function
func processOtp(imagePath string) (string, int64, string, error) {
	// Parse the QR code image
	u, err := parseOtpFromImage(imagePath)
	if err != nil {
		return "", 0, "", fmt.Errorf("failed to parse QR code: %w", err)
	}

	// Extract the secret
	secret := u.Query().Get("secret")
	if secret == "" {
		return "", 0, "", fmt.Errorf("no secret found in the otpauth URL")
	}

	// Extract the period
	period := extractPeriod(u)

	// Extract the issuer and label
	issuer, label := extractIssuerAndLabel(u)
	var provider string
	if issuer != "" && label != "" && !strings.EqualFold(issuer, label) {
		provider = fmt.Sprintf("%s (%s)", issuer, label)
	} else {
		provider = issuer
	}

	return secret, period, provider, nil
}

// generateTOTP generates a TOTP code based on the secret and period
func generateTOTP(secret string, period int64, now time.Time) (string, error) {
	return totp.GenerateCodeCustom(secret, now, totp.ValidateOpts{
		Period:    uint(period),
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
		Skew:      0,
	})
}

// parseOtpFromImage parses the QR code image and extracts the otpauth URL
func parseOtpFromImage(path string) (*url.URL, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	codes, err := goqr.Recognize(img)
	if err != nil {
		return nil, fmt.Errorf("failed to recognize QR code: %w", err)
	}
	if len(codes) == 0 {
		return nil, fmt.Errorf("no QR code found in the image")
	}

	qrData := string(codes[0].Payload)
	if !strings.HasPrefix(qrData, "otpauth://") {
		return nil, fmt.Errorf("QR code does not contain an otpauth URL")
	}

	u, err := url.Parse(qrData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse otpauth URL: %w", err)
	}

	return u, nil
}

// extractPeriod extracts the period from the otpauth URL
func extractPeriod(u *url.URL) int64 {
	period := int64(30) // Default period
	if p := u.Query().Get("period"); p != "" {
		if parsed, err := strconv.ParseInt(p, 10, 64); err == nil && parsed > 0 {
			period = parsed
		}
	}
	return period
}

// extractIssuerAndLabel extracts the issuer and label from the otpauth URL
func extractIssuerAndLabel(u *url.URL) (string, string) {
	path := strings.TrimPrefix(u.Path, "/")
	path, _ = url.QueryUnescape(path)

	issuer := u.Query().Get("issuer")
	if issuer == "" {
		if i := strings.Index(path, ":"); i != -1 {
			issuer = path[:i]
		} else {
			issuer = path
		}
	}
	issuer, _ = url.QueryUnescape(issuer)

	var label string
	if i := strings.Index(path, ":"); i != -1 {
		label = path[i+1:]
	} else {
		label = path
	}
	label = strings.TrimSpace(label)

	return issuer, label
}
