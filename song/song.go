package song

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gopxl/beep/flac"
	"github.com/gopxl/beep/speaker"
	"github.com/gopxl/beep/v2"
	"go.senan.xyz/taglib"
)

type Song struct {
	//tags
	album       string
	artist      string
	title       string
	date        string
	trackNumber string
	genre       []string

	//properties
	trackLength int

	songPath string

	// playing     bool
	// in_playlist map[int]struct{}
}

func NewSong(path string) *Song {
	tags, err := taglib.ReadTags(path)
	if err != nil {
		log.Fatalf("Error parsing metadata: %v", err)
	}
	properties, err := taglib.ReadProperties(path)
	if err != nil {
		log.Fatalf("Error parsing properties: %v", err)

	}

	return &Song{
		album:       firstTag(tags, taglib.Album),
		artist:      firstTag(tags, taglib.Artist),
		title:       firstTag(tags, taglib.Title),
		date:        firstTag(tags, taglib.ReleaseDate),
		trackNumber: firstTag(tags, taglib.TrackNumber),
		genre:       tags[taglib.Genre],
		trackLength: int(properties.SampleRate),

		songPath: path,
	}
}

func firstTag(tags map[string][]string, key string) string {
	if v, ok := tags[key]; ok && len(v) > 0 {
		return v[0]
	}
	return ""
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
	fmt.Printf("Album:   %v\n", s.album)
	fmt.Printf("Artist:   %v\n", s.artist)
	fmt.Printf("Title:   %v\n", s.title)
	fmt.Printf("Length:   %v\n", s.trackLength)

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
