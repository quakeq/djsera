package main

import (
	"example.com/main/song"
)

func main() {
	song := song.NewSong("/run/media/thomas/Extreme SSD/Music/Magdalena Bay - Mercurial World (Deluxe) (2022)/17. All You Do.flac")

	song.PlaySong()
}
