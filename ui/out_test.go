package ui

import (
	"strings"
	"testing"

	"gotest.tools/assert"
)

func TestBox(t *testing.T) {
	out := box("a │ ok │ bee")
	expected := strings.TrimLeft(`
┌───┬────┬─────┐
│ a │ ok │ bee │
└───┴────┴─────┘
`, "\n")
	assert.Equal(t, len(expected), len(out))
	assert.Equal(t, expected, out)
}

func TestSectionWidths(t *testing.T) {
	var testcases = []struct {
		msg      string
		expected []int
	}{
		{
			msg:      "│ one │ two │ three │",
			expected: []int{5, 5, 7},
		},
		{
			msg:      " one │ two │ three ",
			expected: []int{5, 5, 7},
		},
	}

	for _, testcase := range testcases {
		sections := sectionWidths(testcase.msg)
		assert.DeepEqual(t, testcase.expected, sections)
	}
}
