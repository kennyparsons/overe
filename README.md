# Speech-to-Text Go Application

This is a simple Go application that runs in the background, listens for a global hotkey (Cmd+Shift+Space on macOS), records a short audio clip, transcribes it using an external binary, and copies the result to the clipboard.

## Prerequisites

### 1. Go

Make sure you have Go installed on your system. You can download it from [https://golang.org/](https://golang.org/).

### 2. PortAudio

The audio recording functionality depends on the PortAudio library. You'll need to install it on your system.

**On macOS (using Homebrew):**

```bash
brew install portaudio pkg-config
```

### 3. Transcription Binary

This application requires an external binary for audio transcription. You need to specify the path to this binary in the `main.go` file.

Open `main.go` and change the `transcriptionBinaryPath` constant to the correct path:

```go
const (
	// ...
	transcriptionBinaryPath = "/path/to/your/transcription/binary"
)
```

## Setup

1.  **Install dependencies:**

    ```bash
    go mod tidy
    ```

2.  **Build the application:**

    ```bash
    go build
    ```

## Usage

1.  **Run the application:**

    ```bash
    ./stt-app
    ```

2.  **Press `Cmd+Shift+Space`** to start recording. The application will record for 3 seconds.

3.  After the recording is finished, it will be passed to your transcription binary.

4.  If the transcription is successful, the output will be copied to your clipboard, and you'll receive a notification.

5.  If the transcription fails, it will be retried a few times. If it still fails, you'll receive an error notification.

6.  **Press `Ctrl+C`** in the terminal to stop the application.
