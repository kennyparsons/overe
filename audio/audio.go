package audio

import (
	"encoding/binary"
	"os"

	"github.com/gordonklaus/portaudio"
)

const (
	sampleRate = 16000
	channels   = 1
)

func RecordAudio(filePath string, stop <-chan struct{}) error {
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
