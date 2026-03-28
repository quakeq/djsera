package main

import (
	"fmt"
	"os"

	"github.com/gopxl/beep/v2"
	"github.com/quakeq/djsera/filetree"
	"github.com/quakeq/djsera/song"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func main() {

	song.InitSpeaker(beep.SampleRate(44100))
	// Don't pass rows here — songs aren't parsed yet
	t := table.New(
		table.WithColumns(filetree.GetColumns()),
		table.WithFocused(true),
		table.WithHeight(7),
		table.WithWidth(42),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	// NewFiletree parses the dir and calls t.SetRows() internally
	ft := filetree.NewFiletree(t)

	if _, err := tea.NewProgram(ft).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
