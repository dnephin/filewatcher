package files

import (
	"path/filepath"
	"strings"
)

// ExcludeList is a list of file patterns which can be matched against files
type ExcludeList struct {
	patterns []string
}

const (
	anyPath        = "."
	allDirectories = "**" + separator
)

var defaultExclude = []string{
	filepath.Join("**", ".?*"),
	filepath.Join("**", "*~"),
	filepath.Join("**", "*___jb_tmp___"),
	filepath.Join("**", "*___jb_old___"),
}

// NewExcludeList creates a new ExcludeList
func NewExcludeList(patterns []string) (*ExcludeList, error) {
	patterns = append(patterns, defaultExclude...)
	for _, exclude := range patterns {
		if _, err := filepath.Match(exclude, anyPath); err != nil {
			return nil, err
		}
	}
	return &ExcludeList{patterns: patterns}, nil
}

// IsMatch returns true when the filename matches any of the patterns
func (el *ExcludeList) IsMatch(filename string) bool {
	for _, pattern := range el.patterns {
		if matchPath(pattern, filename) {
			return true
		}
	}
	return false
}

func (el *ExcludeList) String() string {
	return strings.Join(el.patterns, ", ")
}

func matchPath(pattern, filename string) bool {
	// patterns were already validated in NewExcludeList, so error
	// can be ignored here
	if match, _ := filepath.Match(pattern, filename); match {
		return true
	}
	if !strings.HasPrefix(pattern, allDirectories) {
		return false
	}

	pattern = strings.TrimPrefix(pattern, allDirectories)
	dirs := splitDirs(filename)
	for i := range dirs {
		if matchPath(pattern, filepath.Join(dirs[i:]...)) {
			return true
		}
	}
	return false
}
