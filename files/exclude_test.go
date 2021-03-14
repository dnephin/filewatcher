package files

import (
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
)

func TestMatchPath(t *testing.T) {
	var testcases = []struct {
		pattern  string
		filename string
		expected bool
	}{
		{pattern: "*/*", filename: "foo/thing/file"},
		{pattern: "**/*", filename: "a/b/file/file.go", expected: true},
		{pattern: "**/*.go", filename: "file"},
		{pattern: "**/*.go", filename: "file/something.txt"},
		{pattern: "**/*.go", filename: "something.go/other.txt"},
		{pattern: "**/bogus", filename: "a/bogus/b"},
		{pattern: "**/*.go", filename: "file.go", expected: true},
		{pattern: "**/*.go", filename: "a/file.go", expected: true},
		{pattern: "**/*.go", filename: "a/b/file.go", expected: true},
		{pattern: "**/*.go", filename: "a/b/c/.go", expected: true},
		{pattern: "**/*.go", filename: "a/b/something.go/next.go", expected: true},
		{pattern: "**/file/*.go", filename: "file/file.go", expected: true},
		{pattern: "**/file/*.go", filename: "a/file/file.go", expected: true},
		{pattern: "**/file/*.go", filename: "a/b/file/file.go", expected: true},
	}
	for _, testcase := range testcases {
		name := fmt.Sprintf(`matchPath("%s","%s")`, testcase.pattern, testcase.filename)
		t.Run(name, func(t *testing.T) {
			actual := matchPath(testcase.pattern, testcase.filename)
			assert.Assert(t, actual == testcase.expected)
		})
	}
}
