package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/atotto/clipboard"
)

// Restore the clipboard content
func restoreClipboard(originalClipboard string) {
	if err := clipboard.WriteAll(originalClipboard); err != nil {
		log.Printf("Warning: could not restore clipboard: %v", err)
	} else {
		fmt.Println("\nOriginal clipboard restored.")
	}
}

// Manage clipboard content and handle signals for graceful exit
func ManageClipboardAndSignals() func() {
	// Save the current clipboard content
	originalClipboard, err := clipboard.ReadAll()
	if err != nil {
		log.Printf("Warning: could not read clipboard to preserve: %v", err)
		originalClipboard = ""
	}

	// Handle signals (SIGINT and SIGTERM) to restore clipboard and exit
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\nRestoring original clipboard and exiting.")
		restoreClipboard(originalClipboard)
		os.Exit(0)
	}()

	// Return cleanup function to restore the clipboard on normal exit
	return func() {
		restoreClipboard(originalClipboard)
	}
}
