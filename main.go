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

	// Determine period
	period := int64(30) // default
	if p := u.Query().Get("period"); p != "" {
		if parsed, err := strconv.ParseInt(p, 10, 64); err == nil && parsed > 0 {
			period = parsed
		}
	}

	issuer, label := extractIssuerAndLabel(u)

	if issuer != "" && label != "" && !strings.EqualFold(issuer, label) {
		fmt.Printf("Provider: %s (%s)\n", issuer, label)
	} else {
		fmt.Printf("Provider: %s\n", issuer)
	}

	var lastCode string
	var lastInterval int64 = -1
	intervalCount := 0

	for {
		now := time.Now()
		interval := now.Unix() / period

		if interval != lastInterval {
			code, err := totp.GenerateCodeCustom(secret, now, totp.ValidateOpts{
				Period:    uint(period),
				Digits:    otp.DigitsSix,
				Algorithm: otp.AlgorithmSHA1,
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

			if intervalCount >= 3 {
				break
			}
		} else {
			remaining := period - (now.Unix() % period)
			fmt.Printf("\rCurrent TOTP code: %s | Expires in: %2d sec", lastCode, remaining)
		}

		time.Sleep(1 * time.Second)
	}

	fmt.Println("\nDone.")
}

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
