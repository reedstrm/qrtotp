package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

type Config struct {
	ImagePath string
}

func InitConfig() *Config {
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
		fmt.Printf("auth version: %s\n", Version)
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
