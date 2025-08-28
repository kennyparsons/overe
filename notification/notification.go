package notification

import "github.com/gen2brain/beeep"

func SendSuccessNotification() {
	beeep.Notify("Transcription Complete", "The transcription has been copied to your clipboard.", "")
}

func SendErrorNotification(err error) {
	beeep.Alert("Transcription Error", err.Error(), "")
}
