package filetree

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/quakeq/djsera/song"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("240"))

type Filetree struct {
	Songs    []*song.Song
	table    table.Model
	MusicDir string
}

func (m Filetree) Init() tea.Cmd { return nil }

func (m Filetree) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			m.PlaySelected(m.table.Cursor())
			return m, tea.Batch()
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m Filetree) View() tea.View {
	return tea.NewView(baseStyle.Render(m.table.View()) + "\n  " + m.table.HelpView() + "\n")
}

func NewFiletree(table table.Model) *Filetree {
	return &Filetree{
		MusicDir: "~/Music",
		table:    table,
	}
}

func GetColumns() []table.Column {
	return []table.Column{
		{Title: "Title", Width: 10},
	}
}

func GetRows() []table.Row {
	return []table.Row{
		{"hi"},
		{"hi2"},
		{"MagBay"},
	}
}

func (f Filetree) parseDir() {

	filepath.WalkDir(f.MusicDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			fmt.Println("[DIR] ", path)
		} else {
			f.Songs = append(f.Songs, song.NewSong(path))
		}
		return nil
	})
}

func (f Filetree) SetMusicDir(dir string) {
	f.MusicDir = dir
	f.parseDir()
}

func (f Filetree) PlaySelected(cursor int) {
	f.Songs[cursor].PlaySong()
}
