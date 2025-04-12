package main

import (
	"fmt"
	"log"
	"time"

	"github.com/atotto/clipboard"
)

var Version = "dev"

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile) // Include timestamps and file info in logs

	// Initialize configuration
	config := InitConfig()

	// Process the image file and extract OTP details
	secret, period, provider, err := ProcessOtp(config.ImagePath)
	if err != nil {
		log.Fatalf("Error processing OTP: %v", err)
	}

	// Handle piped output or interactive mode
	if IsPipedOutput() {
		printOneShotCode(secret, period)
		return
	}

	// Print the provider information
	if provider != "" {
		fmt.Printf("Provider: %s\n", provider)
	}

	interactiveMode(secret, period)
}

// --- Function: one-shot (quiet) mode ---
func printOneShotCode(secret string, period int64) {
	code, err := GenerateTOTP(secret, period, time.Now())
	if err != nil {
		log.Fatalf("Error generating TOTP: %v", err)
	}
	fmt.Println(code)
}

// --- Function: interactive live mode ---
func interactiveMode(secret string, period int64) {
	// Manage clipboard and signal handling
	restoreClipboard := ManageClipboardAndSignals()
	defer restoreClipboard()

	var lastCode string
	var lastInterval int64 = -1
	intervalCount := 0
	var exitAfterInterval int64 = -1

	// Create a ticker that ticks every second
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		interval := now.Unix() / period

		if interval != lastInterval {
			code, err := GenerateTOTP(secret, period, now)
			if err != nil {
				log.Fatalf("Failed to generate TOTP: %v", err)
			}

			err = clipboard.WriteAll(code)
			if err != nil {
				log.Printf("Failed to copy to clipboard: %v", err)
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
	}
}
