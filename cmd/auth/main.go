package main

import (
	"flag"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
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

var version = "dev"

type Config struct {
	ImagePath string
}

func initConfig() *Config {
	config := &Config{}

	// Define flags
	showHelp := flag.Bool("help", false, "Show help message")
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Usage = func() {
		fmt.Println(`Usage: auth <image_file>

This tool extracts TOTP codes from otpauth:// QR images.
It supports both interactive mode (live countdown) and one-shot mode for scripting.

Options:
  --help       Show this help message
  --version    Show version information

⚠️ QR codes contain unencrypted secrets. See SECURITY.md for important usage advice.`)
	}

	// Parse flags
	flag.Parse()

	// Handle --help flag
	if *showHelp {
		flag.Usage()
		os.Exit(0)
	}

	// Handle --version flag
	if *showVersion {
		fmt.Printf("auth version: %s\n", version)
		os.Exit(0)
	}

	// Handle positional argument for the image file
	args := flag.Args()
	if len(args) < 1 {
		log.Println("Error: <image_file> argument is required")
		flag.Usage()
		os.Exit(1)
	}
	config.ImagePath = args[0]

	return config
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile) // Include timestamps and file info in logs

	// Initialize configuration
	config := initConfig()

	// Process the image file
	u, err := parseOtpFromImage(config.ImagePath)
	if err != nil {
		log.Fatalf("Error Parsing QR code: %v", err)
	}

	secret := u.Query().Get("secret")
	if secret == "" {
		log.Println("No secret found in the otpauth URL.")
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

// --- Function: Generate TOTP ---
func generateTOTP(secret string, period int64, now time.Time) (string, error) {
	return totp.GenerateCodeCustom(secret, now, totp.ValidateOpts{
		Period:    uint(period),
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
		Skew:      0,
	})
}

// --- Function: one-shot (quiet) mode ---
func printOneShotCode(secret string, period int64) {
	code, err := generateTOTP(secret, period, time.Now())
	if err != nil {
		log.Fatalf("Error generating TOTP: %v", err)
	}
	fmt.Println(code)
}

// Helper function to restore the clipboard
func restoreClipboard(originalClipboard string) {
	if err := clipboard.WriteAll(originalClipboard); err != nil {
		log.Printf("Warning: could not restore clipboard: %v", err)
	} else {
		fmt.Println("\nOriginal clipboard restored.")
	}
}

// Helper function to manage clipboard and signal handling
func manageClipboardAndSignals() {
	// Save the current clipboard content
	originalClipboard, err := clipboard.ReadAll()
	if err != nil {
		log.Printf("Warning: could not read clipboard to preserve: %v", err)
		originalClipboard = ""
	}

	// Defer clipboard restoration
	defer restoreClipboard(originalClipboard)

	// Handle Ctrl+C signal to restore clipboard and exit
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\nRestoring original clipboard and exiting.")
		restoreClipboard(originalClipboard)
		os.Exit(0)
	}()
}

// --- Function: interactive live mode ---
func interactiveMode(secret string, period int64) {
	// Manage clipboard and signal handling
	manageClipboardAndSignals()

	var lastCode string
	var lastInterval int64 = -1
	intervalCount := 0
	var exitAfterInterval int64 = -1

	for {
		now := time.Now()
		interval := now.Unix() / period

		if interval != lastInterval {
			code, err := generateTOTP(secret, period, now)
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

		time.Sleep(1 * time.Second)
	}
}

// --- Function: detect pipe vs terminal ---
func isPipedOutput() bool {
	info, err := os.Stdout.Stat()
	if err != nil {
		log.Printf("Error checking output mode: %v", err)
		return false
	}
	return (info.Mode() & os.ModeCharDevice) == 0
}
