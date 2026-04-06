package filetree

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/quakeq/djsera/song"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("240"))

type Model struct {
	Songs    []song.Song
	cursor   int
	MusicDir string
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if m.cursor < len(m.Songs)-1 {
				m.cursor++
			}
		case "enter":
			cmd = PlaySongCmd(m.Songs[m.cursor])
			return m, cmd
		}
	}
	return m, cmd
}

func (m Model) View() tea.View {
	// The header
	s := "Songs\n\n"

	// Iterate over our choices
	for i, song := range m.Songs {

		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
		}

		// Render the row
		s += fmt.Sprintf("%s %s   %s \n", cursor, song.Title, song.Artist)
	}

	// The footer
	s += "\nPress q to quit.\n"

	view := tea.NewView(s)
	view.AltScreen = true
	return view
}

func NewModel() *Model {
	home, _ := os.UserHomeDir()
	musicDir := filepath.Join(home, "Music")
	songs := ParseDir(musicDir)

	return &Model{
		MusicDir: musicDir,
		Songs:    songs,
	}
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

func (f *Model) SetMusicDir(dir string) {
	f.MusicDir = dir
	f.Songs = ParseDir(f.MusicDir)
}

func PlaySongCmd(s song.Song) tea.Cmd {
	return func() tea.Msg {
		s.PlaySong()
		return nil
	}
}
