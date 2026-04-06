package main

import (
	"fmt"
	"os"

	"github.com/gopxl/beep/v2"
	"github.com/quakeq/djsera/filetree"
	"github.com/quakeq/djsera/song"

	tea "charm.land/bubbletea/v2"
)

func main() {

	song.InitSpeaker(beep.SampleRate(44100))

	// NewFiletree parses the dir and calls t.SetRows() internally
	ft := filetree.NewModel()

	if _, err := tea.NewProgram(ft).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
