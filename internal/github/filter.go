package github

import "strings"

var skipExt = []string{
	".json", ".lock", ".sum", ".yaml",
	".yml", ".md", ".txt",
}

var allowExt = []string{
	".go", ".js", ".ts", ".py",
}

func IsReviewable(f PRFile) bool {

	for _, s := range skipExt {
		if strings.HasSuffix(f.Filename, s) {
			return false
		}
	}

	for _, a := range allowExt {
		if strings.HasSuffix(f.Filename, a) {
			return true
		}
	}

	return false
}
