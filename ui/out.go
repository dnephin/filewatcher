package ui

import (
	"bytes"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/fatih/color"
)

const (
	pipe        = '│'
	bar         = '─'
	leftCorner  = 0
	rightCorner = 1
	tee         = 2
)

var (
	top    = []rune{'┌', '┐', '┬'}
	bottom = []rune{'└', '┘', '┴'}
)

// PrintStart message to inform the user that a process is being executed
func PrintStart(cmd []string) {
	msg := "filewatcher │ " + strings.Join(cmd, " ")
	fmt.Print(color.YellowString(box(msg)))
}

// PrintEnd message to inform the user that a process is done
func PrintEnd(elapsed time.Duration, filename string, err error) {
	msg := "OK"
	msgColor := color.GreenString
	if err != nil {
		msg = err.Error()
		msgColor = color.HiRedString
	}
	out := fmt.Sprintf("%s │ %s │ %s", msg, filename, elapsed)
	fmt.Print(msgColor(box(out)))
}

func box(msg string) string {
	buf := new(bytes.Buffer)
	msg = " " + msg + " "
	sections := sectionWidths(msg)

	horizontal(buf, top, sections)
	buf.WriteString(s(pipe) + msg + s(pipe) + "\n")

	horizontal(buf, bottom, sections)
	return buf.String()
}

func s(r rune) string {
	return string(r)
}

func sectionWidths(msg string) []int {
	sections := []int{}

	for {
		i := strings.IndexRune(msg, pipe)
		switch {
		case i < 0:
			end := utf8.RuneCountInString(msg)
			if end == 0 {
				return sections
			}
			return append(sections, end)
		case i > 0:
			sections = append(sections, utf8.RuneCountInString(msg[:i]))
		}
		msg = msg[i+utf8.RuneLen(pipe):]
	}
}

func horizontal(buf *bytes.Buffer, runes []rune, sections []int) {
	buf.WriteRune(runes[leftCorner])
	last := len(sections) - 1
	for i, width := range sections {
		buf.WriteString(strings.Repeat(string(bar), width))
		if i != last {
			buf.WriteRune(runes[tee])
		}
	}
	buf.WriteString(s(runes[rightCorner]) + "\n")
}
