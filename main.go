package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/gen2brain/beeep"
	"github.com/gordonklaus/portaudio"
	"golang.design/x/hotkey"
	"golang.design/x/hotkey/mainthread"
)

const (
	transcriptionRetries    = 3
	transcriptionBinaryPath = "/opt/homebrew/bin/go-transcribe"
	sampleRate              = 16000
	channels                = 1
)

func main() {
	// Set up logging to a file in the user's log directory.
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Could not get user home directory: %v", err)
	}
	logDir := filepath.Join(homeDir, "Library", "Logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Fatalf("Could not create log directory: %v", err)
	}
	logFilePath := filepath.Join(logDir, "stt-app.log")
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Could not open log file: %v", err)
	}
	defer logFile.Close()
	// We use io.MultiWriter to log to both stdout and the log file.
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))

	mainthread.Init(run)
}

func run() {
	log.Println("Starting speech-to-text application...")
	log.Println("Press Ctrl+Shift+Space to start recording.")
	log.Println("Press Ctrl+C in the terminal to exit.")

	hk := hotkey.New([]hotkey.Modifier{hotkey.ModCtrl, hotkey.ModShift}, hotkey.KeySpace)
	if err := hk.Register(); err != nil {
		log.Fatalf("hotkey: failed to register hotkey: %v", err)
	}
	log.Printf("hotkey: %v is registered", hk)

	doneChan := make(chan bool)

	// This is the main, synchronous event loop.
	for {
		// 1. Block and wait for a key press.
		<-hk.Keydown()

		// 2. Handle the entire recording process in a separate goroutine.
		go handleHotkey(doneChan)

		// 3. Block the main loop, making it deaf to new key presses until the
		//    entire process is complete.
		<-doneChan

		// 4. DRAIN PHASE: After the process is done, we aggressively drain any
		//    phantom keydown events that may have been queued by the OS while
		//    the AppleScript dialog was open.
	DrainLoop:
		for {
			select {
			case <-hk.Keydown():
				// An event was queued; discard it and check again.
				log.Println("Drained a phantom keydown event.")
			default:
				// The channel is empty; we are safe to listen for new events.
				break DrainLoop
			}
		}
	}
}

func handleHotkey(doneChan chan bool) {
	// Ensure we signal the main loop to continue when this function exits.
	defer func() { doneChan <- true }()

	log.Println("Hotkey pressed. Starting recording...")

	tempFile, err := os.CreateTemp("", "stt-recording-*.wav")
	if err != nil {
		log.Printf("Error creating temp file: %v", err)
		sendErrorNotification(err)
		return
	}
	// defer os.Remove(tempFile.Name())
	tempFilePath, _ := filepath.Abs(tempFile.Name())
	log.Printf("Recording audio to temporary file: %s", tempFilePath)
	tempFile.Close()

	stopChan := make(chan struct{})
	recordingErrChan := make(chan error, 1)

	go func() {
		recordingErrChan <- recordAudio(tempFilePath, stopChan)
	}()

	shouldTranscribe := showAppleScriptPopup()

	close(stopChan) // Signal recording to stop

	if err := <-recordingErrChan; err != nil {
		log.Printf("Error recording audio: %v", err)
		sendErrorNotification(err)
		return
	}

	if shouldTranscribe {
		log.Println("Recording finished.")
		transcribe(tempFilePath)
	} else {
		log.Println("Recording canceled.")
	}
}

func showAppleScriptPopup() bool {
	script := `display dialog "Recording in progress..." buttons {"Cancel", "Stop & Transcribe"} default button "Stop & Transcribe"`
	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.CombinedOutput()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 && strings.Contains(string(output), "User canceled") {
			return false
		}
		log.Printf("Error showing AppleScript dialog: %v, output: %s", err, string(output))
		return false
	}
	return strings.Contains(string(output), "Stop & Transcribe")
}

func recordAudio(filePath string, stop <-chan struct{}) error {
	portaudio.Initialize()
	defer portaudio.Terminate()

	buffer := make([]int16, 0)
	stream, err := portaudio.OpenDefaultStream(channels, 0, float64(sampleRate), 1024, func(in []int16) {
		buffer = append(buffer, in...)
	})
	if err != nil {
		return err
	}
	defer stream.Close()

	if err := stream.Start(); err != nil {
		return err
	}
	defer stream.Stop()

	<-stop // Wait until the stop channel is closed.

	return writeWav(filePath, buffer, channels, sampleRate, 16)
}

func transcribe(audioFilePath string) {
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
		sendErrorNotification(fmt.Errorf("transcription failed after %d attempts: %v", transcriptionRetries, string(output)))
		return
	}

	transcriptionContent, readErr := os.ReadFile(transcriptionFilePath)
	if readErr != nil {
		log.Printf("Error reading transcription file: %v", readErr)
		sendErrorNotification(fmt.Errorf("could not read transcription file: %v", readErr))
		return
	}

	transcription := string(transcriptionContent)
	if err := clipboard.WriteAll(transcription); err != nil {
		log.Printf("Error copying to clipboard: %v", err)
		sendErrorNotification(err)
		return
	}

	log.Println("Transcription successful.")
	log.Printf("Copied to clipboard: %s", transcription)
	sendSuccessNotification()
}

func sendSuccessNotification() {
	beeep.Notify("Transcription Complete", "The transcription has been copied to your clipboard.", "")
}

func sendErrorNotification(err error) {
	beeep.Alert("Transcription Error", err.Error(), "")
}

func writeWav(path string, buffer []int16, channels, sampleRate, bitDepth int) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	file.WriteString("RIFF")
	file.Write(make([]byte, 4))
	file.WriteString("WAVE")
	file.WriteString("fmt ")
	binary.Write(file, binary.LittleEndian, uint32(16))
	binary.Write(file, binary.LittleEndian, uint16(1))
	binary.Write(file, binary.LittleEndian, uint16(channels))
	binary.Write(file, binary.LittleEndian, uint32(sampleRate))
	binary.Write(file, binary.LittleEndian, uint32(sampleRate*channels*bitDepth/8))
	binary.Write(file, binary.LittleEndian, uint16(channels*bitDepth/8))
	binary.Write(file, binary.LittleEndian, uint16(bitDepth))
	file.WriteString("data")
	binary.Write(file, binary.LittleEndian, uint32(len(buffer)*2))

	for _, v := range buffer {
		binary.Write(file, binary.LittleEndian, v)
	}

	fileSize, _ := file.Seek(0, 2)
	file.Seek(4, 0)
	binary.Write(file, binary.LittleEndian, uint32(fileSize-8))

	return nil
}
