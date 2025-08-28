package ui

import (
	"log"
	"os/exec"
	"strings"
)

func ShowAppleScriptPopup() bool {
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
