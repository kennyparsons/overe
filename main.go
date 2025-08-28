package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"stt-app/hotkey"
	"golang.design/x/hotkey/mainthread"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "version":
			fmt.Printf("overe version: %s\n", version)
			return
		case "run":
			// continue to normal execution
		default:
			fmt.Printf("Unknown command: %s\n", os.Args[1])
			fmt.Println("Available commands: run, version")
			os.Exit(1)
		}
	}

	// Set up logging to a file in the user's log directory.
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Could not get user home directory: %v", err)
	}

	logDir := filepath.Join(homeDir, "Library", "Logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Fatalf("Could not create log directory: %v", err)
	}
	logFilePath := filepath.Join(logDir, "overe.log")
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Could not open log file: %v", err)
	}
	defer logFile.Close()
	// We use io.MultiWriter to log to both stdout and the log file.
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))

	mainthread.Init(hotkey.Run)
}