package hotkey

import (
	"log"
	"os"
	"path/filepath"

	"stt-app/audio"
	"stt-app/notification"
	"stt-app/transcription"
	"stt-app/ui"
	"golang.design/x/hotkey"
)

func Run() {
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
		notification.SendErrorNotification(err)
		return
	}
	// defer os.Remove(tempFile.Name())
	tempFilePath, _ := filepath.Abs(tempFile.Name())
	log.Printf("Recording audio to temporary file: %s", tempFilePath)
	tempFile.Close()

	stopChan := make(chan struct{})
	recordingErrChan := make(chan error, 1)

	go func() {
		recordingErrChan <- audio.RecordAudio(tempFilePath, stopChan)
	}()

	shouldTranscribe := ui.ShowAppleScriptPopup()

	close(stopChan) // Signal recording to stop

	if err := <-recordingErrChan; err != nil {
		log.Printf("Error recording audio: %v", err)
		notification.SendErrorNotification(err)
		return
	}

	if shouldTranscribe {
		log.Println("Recording finished.")
		transcription.Transcribe(tempFilePath)
	} else {
		log.Println("Recording canceled.")
	}
}
