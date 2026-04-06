package song

import (
	// "fmt"
	"log"
	"os"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/flac"
	"github.com/gopxl/beep/v2/speaker"
	"go.senan.xyz/taglib"
)

const sampleRate = beep.SampleRate(44100)

type Song struct {
	//tags
	Album       string
	Artist      string
	Title       string
	Data        string
	TrackNumber string
	Genre       []string

	//properties
	TrackLength int

	songPath string

	// playing     bool
	// in_playlist map[int]struct{}
}

func NewSong(path string) Song {
	tags, err := taglib.ReadTags(path)
	if err != nil {
		log.Fatalf("Error parsing metadata: %v", err)
	}
	properties, err := taglib.ReadProperties(path)
	if err != nil {
		log.Fatalf("Error parsing properties: %v", err)

	}

	return Song{
		Album:       firstTag(tags, taglib.Album),
		Artist:      firstTag(tags, taglib.Artist),
		Title:       firstTag(tags, taglib.Title),
		Data:        firstTag(tags, taglib.ReleaseDate),
		TrackNumber: firstTag(tags, taglib.TrackNumber),
		Genre:       tags[taglib.Genre],
		TrackLength: int(properties.Length.Round(time.Second).Seconds()),

		songPath: path,
	}
}

func InitSpeaker(sampleRate beep.SampleRate) {
	speaker.Init(sampleRate, sampleRate.N(time.Second/10))
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

	var finalStreamer beep.Streamer
	if format.SampleRate != sampleRate {
		finalStreamer = beep.Resample(4, format.SampleRate, sampleRate, streamer)
	} else {
		finalStreamer = streamer
	}
	streamer.Len()

	speaker.Clear()

	// Play the stream and wait until it finishes
	done := make(chan bool)
	speaker.Play(beep.Seq(finalStreamer, beep.Callback(func() {
		done <- true
	})))
	<-done
}
