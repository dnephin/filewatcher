package files

import (
	"testing"

	a "github.com/stretchr/testify/assert"
)

func TestIsAnyDirMatchNotAllDirPrefix(t *testing.T) {
	a.False(t, isAnyDirMatch("*/*", "./file"))
}

func TestIsAnyDirMatchNotAMatch(t *testing.T) {
	a.False(t, isAnyDirMatch("**/*.go", "file"))
	a.False(t, isAnyDirMatch("**/*.go", "file/something.txt"))
	a.False(t, isAnyDirMatch("**/*.go", "something.go/other.txt"))
	a.False(t, isAnyDirMatch("**/bogus", "a/bogus/b"))
}

func TestIsAnyDirMatch(t *testing.T) {
	a.True(t, isAnyDirMatch("**/*.go", "file.go"))
	a.True(t, isAnyDirMatch("**/*.go", "a/file.go"))
	a.True(t, isAnyDirMatch("**/*.go", "a/b/file.go"))
	a.True(t, isAnyDirMatch("**/*.go", "a/b/c/.go"))
	a.True(t, isAnyDirMatch("**/file/*.go", "file/file.go"))
	a.True(t, isAnyDirMatch("**/file/*.go", "a/file/file.go"))
	a.True(t, isAnyDirMatch("**/file/*.go", "a/b/file/file.go"))
}
