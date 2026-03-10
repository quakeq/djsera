package song

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gopxl/beep/flac"
	"github.com/gopxl/beep/speaker"
	"github.com/gopxl/beep/v2"
)

type Song struct {
	// album  string
	// artist string
	// title  string
	// date   string

	songPath string

	// trackLength float32
	// trackNumber int
	// totalTracks int

	// genre []string

	// playing     bool
	// in_playlist map[int]struct{}
}

func NewSong(path string) *Song {
	return &Song{
		songPath: path,
	}
}

func (s Song) PlaySong() {
	songReader, err := os.Open(s.songPath)

	if err != nil {
		log.Fatalf("Error opening FLAC: %v", err)
	}

	// Decode the FLAC stream
	streamer, format, err := flac.Decode(songReader)
	if err != nil {
		log.Fatalf("Error decoding FLAC: %v", err)
	}
	defer streamer.Close()

	fmt.Printf("Playing: %s\n", s.songPath)
	fmt.Printf("Sample Rate: %d Hz\n", format.SampleRate)
	fmt.Printf("Channels:    %d\n", format.NumChannels)
	fmt.Printf("Precision:   %d-bit\n", format.Precision*8)

	// Initialize the speaker with the file's sample rate
	// Buffer size: 1/10th of a second
	bufferSize := format.SampleRate.N(time.Second / 10)
	err = speaker.Init(format.SampleRate, bufferSize)
	if err != nil {
		log.Fatalf("Error initializing speaker: %v", err)
	}

	// Play the stream and wait until it finishes
	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))

	<-done
}
