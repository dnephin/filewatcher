package files

import (
	"os"
	"path/filepath"
	"strings"

	log "github.com/Sirupsen/logrus"
)

const (
	separator = string(filepath.Separator)
)

// SplitDirs splits a path into directory segments after cleaning the path
func SplitDirs(path string) []string {
	return strings.Split(filepath.Clean(path), separator)
}

func isMaxDepth(path string, depth int) bool {
	return len(SplitDirs(path)) == depth
}

// WalkDirectories walks each directory in the slice to the desired depth
// and returns a new slice which contains all the directories walked.
// Directories may be excluded using the exclude slice.
func WalkDirectories(dirs []string, depth int, exclude *ExcludeList) []string {
	output := []string{}

	walker := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Warnf("Error walking '%s': %s", path, err)
			return nil
		}
		if !info.IsDir() {
			return nil
		}
		if isMaxDepth(path, depth) || exclude.IsMatch(path) {
			return filepath.SkipDir
		}

		output = append(output, path)
		return nil
	}

	for _, dir := range dirs {
		filepath.Walk(dir, walker)
	}
	return output
}
