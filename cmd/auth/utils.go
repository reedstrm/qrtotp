package main

import (
	"log"
	"os"
)

// Check if the output is being piped
func isPipedOutput() bool {
	info, err := os.Stdout.Stat()
	if err != nil {
		log.Printf("Error checking output mode: %v", err)
		return false
	}
	return (info.Mode() & os.ModeCharDevice) == 0
}
