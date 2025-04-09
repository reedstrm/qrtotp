package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/liyue201/goqr"
	"github.com/pquerna/otp/totp"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: qrtotp <image_file>")
		os.Exit(1)
	}

	filePath := os.Args[1]
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Failed to open image: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		fmt.Printf("Failed to decode image: %v\n", err)
		os.Exit(1)
	}

	codes, err := goqr.Recognize(img)
	if err != nil {
		fmt.Printf("Failed to recognize QR code: %v\n", err)
		os.Exit(1)
	}
	if len(codes) == 0 {
		fmt.Println("No QR code found in the image.")
		os.Exit(1)
	}

	qrData := string(codes[0].Payload)
	if !strings.HasPrefix(qrData, "otpauth://") {
		fmt.Println("QR code does not contain an otpauth URL.")
		os.Exit(1)
	}

	u, err := url.Parse(qrData)
	if err != nil {
		fmt.Printf("Failed to parse otpauth URL: %v\n", err)
		os.Exit(1)
	}

	secret := u.Query().Get("secret")
	if secret == "" {
		fmt.Println("No secret found in the otpauth URL.")
		os.Exit(1)
	}

	issuer, label := extractIssuerAndLabel(u)

	// Generate TOTP
	passcode, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		fmt.Printf("Failed to generate TOTP: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Provider: %s (%s)\n", issuer, label)
	fmt.Printf("Current TOTP code: %s\n", passcode)
}

func extractIssuerAndLabel(u *url.URL) (string, string) {
	// Path like "/Issuer:Label" or "/Label"
	path := strings.TrimPrefix(u.Path, "/")
	path, _ = url.QueryUnescape(path)

	// Extract issuer
	issuer := u.Query().Get("issuer")
	if issuer == "" {
		if i := strings.Index(path, ":"); i != -1 {
			issuer = path[:i]
		} else {
			issuer = path
		}
	}
	issuer, _ = url.QueryUnescape(issuer)

	// Extract label
	var label string
	if i := strings.Index(path, ":"); i != -1 {
		label = path[i+1:]
	} else {
		label = path
	}
	label = strings.TrimSpace(label)

	return issuer, label
}

