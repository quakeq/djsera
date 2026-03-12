package filetree

import "github.com/quakeq/djsera/song"

type Filetree struct {
	songs    []string
	cursor   int
	selected map[int]struct{}
}

func (f Filetree) AddSongs(songs ...song.Song) {

}
