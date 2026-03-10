package filetree

type Filetree struct {
	choices  []string
	cursor   int
	selected map[int]struct{}
}
