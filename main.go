package main

import (
	"fmt"
	"os"

	"github.com/quakeq/djsera/filetree"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func main() {

	t := table.New(
		table.WithColumns(filetree.GetColumns()),
		table.WithRows(filetree.GetRows()),
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

	filetree := filetree.NewFiletree(t)

	if _, err := tea.NewProgram(filetree).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
