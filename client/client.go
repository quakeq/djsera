package client

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/quakeq/djsera/filetree"
)

type sessionState uint

const (
	filetreeView sessionState = iota
)

var (
	modelStyle = lipgloss.NewStyle().
			Width(15).
			Height(5).
			Align(lipgloss.Center, lipgloss.Center).
			BorderStyle(lipgloss.HiddenBorder())
	focusedModelStyle = lipgloss.NewStyle().
				Width(15).
				Height(5).
				Align(lipgloss.Center, lipgloss.Center).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("69"))
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

type Model struct {
	state sessionState
	index int

	filetree filetree.Model
}

func newModel() Model {
	m := Model{state: filetreeView}
	m.filetree = *filetree.NewModel()
	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.filetree.Init())
}

func (m Model) View() tea.View {
	var s strings.Builder
	model := m.currentFocusedModel()
	if m.state == filetreeView {
		s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, focusedModelStyle.Render(fmt.Sprintf("%4s", m.timer.View())), modelStyle.Render(m.spinner.View())))
	}
	s.WriteString(helpStyle.Render(fmt.Sprintf("\ntab: focus next • n: new %s • q: exit\n", model)))
	return tea.NewView(s.String())
}

func (m Model) currentFocusedModel() string {
	return "filetree"

}

func (m Model) Next() {
	if m.index == len(spinners)-1 {
		m.index = 0
	} else {
		m.index++
	}
}
