package github

type PRFile struct {
	Filename  string `json:"filename"`
	Status    string `json:"status"`
	Patch     string `json:"patch"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
}

type PRMeta struct {
	Number int
	Title  string
	Author string
	Files  []PRFile
}
