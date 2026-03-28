package filetree

import (
	"io/fs"
	"os"
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
	Songs    []song.Song
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
			cmd = PlaySongCmd(m.Songs[m.table.Cursor()])
			m.table, _ = m.table.Update(msg)
			return m, cmd
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m Filetree) View() tea.View {
	return tea.NewView(baseStyle.Render(m.table.View()) + "\n  " + m.table.HelpView() + "\n")
}

func NewFiletree(t table.Model) *Filetree {
	home, _ := os.UserHomeDir()
	musicDir := filepath.Join(home, "Music")
	songs := ParseDir(musicDir)

	t.SetRows(RowsFromSongs(songs))

	return &Filetree{
		MusicDir: musicDir,
		table:    t,
		Songs:    songs,
	}
}

func GetColumns() []table.Column {
	return []table.Column{
		{Title: "Title", Width: 10},
	}
}

func RowsFromSongs(songs []song.Song) []table.Row {
	rows := make([]table.Row, 0, len(songs))
	for _, s := range songs {
		rows = append(rows, table.Row{s.Title}) // adjust fields to match song.Song
	}
	return rows
}

func ParseDir(dir string) []song.Song {
	var songs []song.Song
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			songs = append(songs, song.NewSong(path))
		}
		return nil
	})
	return songs
}

func (f *Filetree) SetMusicDir(dir string) {
	f.MusicDir = dir
	f.Songs = ParseDir(f.MusicDir)
}

func PlaySongCmd(s song.Song) tea.Cmd {
	return func() tea.Msg {
		s.PlaySong()
		return nil
	}
}
