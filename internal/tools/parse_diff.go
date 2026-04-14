package tools

type ChangedLine struct {
	File    string
	Line    int
	Content string
}

type ParseDiffResult struct {
	Files []string
	Lines []ChangedLine
}
