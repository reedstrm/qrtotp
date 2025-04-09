package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/atotto/clipboard"
	"github.com/liyue201/goqr"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: auth <image_file>")
		os.Exit(1)
	}

	u, err := parseOtpFromImage(os.Args[1])
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	secret := u.Query().Get("secret")
	if secret == "" {
		fmt.Println("No secret found in the otpauth URL.")
		os.Exit(1)
	}

	period := extractPeriod(u)

	if isPipedOutput() {
		printOneShotCode(secret, period)
		return
	}

	issuer, label := extractIssuerAndLabel(u)
	if issuer != "" && label != "" && !strings.EqualFold(issuer, label) {
		fmt.Printf("Provider: %s (%s)\n", issuer, label)
	} else {
		fmt.Printf("Provider: %s\n", issuer)
	}

	interactiveMode(secret, period)
}

// --- Function: QR + otpauth URL parsing ---
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

// --- Function: period parsing ---
func extractPeriod(u *url.URL) int64 {
	period := int64(30) // default
	if p := u.Query().Get("period"); p != "" {
		if parsed, err := strconv.ParseInt(p, 10, 64); err == nil && parsed > 0 {
			period = parsed
		}
	}
	return period
}

// --- Function: label + issuer parsing ---
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

// --- Function: one-shot (quiet) mode ---
func printOneShotCode(secret string, period int64) {
	code, err := totp.GenerateCodeCustom(secret, time.Now(), totp.ValidateOpts{
		Period:    uint(period),
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
		Skew:      0,
	})
	if err != nil {
		fmt.Printf("Error generating TOTP: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(code)
}

// --- Function: interactive live mode ---
func interactiveMode(secret string, period int64) {
	// Save and restore clipboard
	originalClipboard, err := clipboard.ReadAll()
	if err != nil {
		fmt.Println("Warning: could not read clipboard to preserve:", err)
		originalClipboard = ""
	}

	// Restore on Ctrl+C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\nRestoring original clipboard and exiting.")
		_ = clipboard.WriteAll(originalClipboard)
		os.Exit(0)
	}()
	// Restore on normal exit
	defer func() {
		_ = clipboard.WriteAll(originalClipboard)
		fmt.Println("\nOriginal clipboard restored.")
	}()

	var lastCode string
	var lastInterval int64 = -1
	intervalCount := 0
	var exitAfterInterval int64 = -1

	for {
		now := time.Now()
		interval := now.Unix() / period

		if interval != lastInterval {
			code, err := totp.GenerateCodeCustom(secret, now, totp.ValidateOpts{
				Period:    uint(period),
				Digits:    otp.DigitsSix,
				Algorithm: otp.AlgorithmSHA1,
				Skew:      0,
			})
			if err != nil {
				fmt.Printf("Failed to generate TOTP: %v\n", err)
				os.Exit(1)
			}

			err = clipboard.WriteAll(code)
			if err != nil {
				fmt.Printf("Failed to copy to clipboard: %v\n", err)
			}

			lastCode = code
			lastInterval = interval
			intervalCount++

			fmt.Printf("\rCurrent TOTP code: %s | Expires in: %2d sec", code, period-(now.Unix()%period))

			if intervalCount == 3 {
				exitAfterInterval = interval + 1
			}
		} else {
			remaining := period - (time.Now().Unix() % period)
			fmt.Printf("\rCurrent TOTP code: %s | Expires in: %2d sec", lastCode, remaining)
		}

		if exitAfterInterval > 0 && interval >= exitAfterInterval {
			fmt.Println("\r\033[KDone.")
			break
		}

		time.Sleep(1 * time.Second)
	}
}

// --- Function: detect pipe vs terminal ---
func isPipedOutput() bool {
	info, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) == 0
}
