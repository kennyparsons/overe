package transcription

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"stt-app/notification"
)

const (
	transcriptionRetries    = 3
	transcriptionBinaryPath = "/opt/homebrew/bin/go-transcribe"
)

func Transcribe(audioFilePath string) {
	transcriptionFilePath := strings.TrimSuffix(audioFilePath, ".wav") + ".txt"
	defer os.Remove(transcriptionFilePath)

	var err error
	var output []byte

	for i := 0; i < transcriptionRetries; i++ {
		log.Printf("Transcription attempt %d/%d", i+1, transcriptionRetries)
		homeDir := os.Getenv("HOME")
		modelPath := filepath.Join(homeDir, ".config/whisper-cpp/models/ggml-medium.en.bin")
		cmd := exec.Command(transcriptionBinaryPath, "--model", modelPath, audioFilePath)
		// When running as a bundled .app, the PATH is not inherited from the shell.
		// We must explicitly provide the path to Homebrew and other required binaries.
		vlcPath := "/Applications/VLC.app/Contents/MacOS"
		homebrewPath := "/opt/homebrew/bin"
		newPath := fmt.Sprintf("PATH=%s:%s:%s", vlcPath, homebrewPath, os.Getenv("PATH"))
		cmd.Env = append(os.Environ(), newPath)
		output, err = cmd.CombinedOutput()
		if err == nil {
			break
		}
		log.Printf("Transcription command failed: %v, output: %s", err, string(output))
		time.Sleep(1 * time.Second)
	}

	if err != nil {
		notification.SendErrorNotification(fmt.Errorf("transcription failed after %d attempts: %v", transcriptionRetries, string(output)))
		return
	}

	transcriptionContent, readErr := os.ReadFile(transcriptionFilePath)
	if readErr != nil {
		log.Printf("Error reading transcription file: %v", readErr)
		notification.SendErrorNotification(fmt.Errorf("could not read transcription file: %v", readErr))
		return
	}

	transcription := string(transcriptionContent)
	if err := clipboard.WriteAll(transcription); err != nil {
		log.Printf("Error copying to clipboard: %v", err)
		notification.SendErrorNotification(err)
		return
	}

	log.Println("Transcription successful.")
	log.Printf("Copied to clipboard: %s", transcription)
	notification.SendSuccessNotification()
}
