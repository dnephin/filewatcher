package files

import (
	"os"
	"path/filepath"
	"strings"

	log "github.com/Sirupsen/logrus"
)

func isMaxDepth(path string, depth int) bool {
	sep := string(filepath.Separator)
	return len(strings.Split(filepath.Clean(path), sep)) == depth
}

// WalkDirectories walks each directory in the slice to the desired depth
// and returns a new slice which contains all the directories walked.
// Directories may be excluded using the exclude slice.
func WalkDirectories(dirs []string, depth int, exclude []string) []string {
	output := []string{}

	walker := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Warnf("Error walking '%s': %s", path, err)
			return nil
		}
		if !info.IsDir() {
			return nil
		}

		if isMaxDepth(path, depth) {
			return filepath.SkipDir
		}

		// TODO: use exclude
		output = append(output, path)
		return nil
	}

	for _, dir := range dirs {
		filepath.Walk(dir, walker)
	}
	return output
}
