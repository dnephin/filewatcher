package ui

import (
	"fmt"
	"io"
	"strings"

	tm "github.com/buger/goterm"
)

const border = "─ │ ┌ ┐ └ ┘"

// UI stores terminal UI elements
type UI struct {
	header *tm.Box
	footer *tm.Box
	logs   *tm.Box
}

// SetupUI elements
func SetupUI() (*UI, error) {
	width := 100 | tm.PCT
	ui := &UI{
		header: tm.NewBox(width, 3, 0),
		logs:   tm.NewBox(width, 3, 0),
		footer: tm.NewBox(width, 3, 0),
	}
	ui.logs.Border = "           "
	ui.logs.PaddingX = 0
	// TODO: remove this once upstream is fixed
	ui.footer.Border = border
	ui.header.Border = border

	ui.draw()
	return ui, nil
}

// Output returns a an io.Writer for writing into the main output section of the UI
func (ui *UI) Output() io.Writer {
	return &flushWriter{Writer: ui.logs, ui: ui}
}

func (ui *UI) draw() {
	tm.Clear()

	height := tm.Height()
	ui.logs.Height = height - 6
	tm.Print(tm.MoveTo(ui.header.String(), 1, 1))
	tm.Print(tm.MoveTo(ui.logs.String(), 1, 3))
	tm.Print(tm.MoveTo(ui.footer.String(), 1, height-2))

	// Hide the cursor
	fmt.Fprint(tm.Screen, "\033[?25l")
	tm.MoveCursor(0, 0)
	tm.Flush()
}

// Header sets the header to the specified string
func (ui *UI) Header(command []string) {
	formatted := formatHeader(command)
	ui.header.Buf.Reset()
	ui.header.Buf.WriteString(formatted)
	ui.draw()
}

// Footer sets the footer to the specified string
func (ui *UI) Footer(footer string) {
	ui.footer.Buf.Reset()
	ui.footer.Buf.WriteString(footer)
	ui.draw()
}

// Reset restores the terminal
func (ui *UI) Reset() {
	// Restore cursor
	// TODO: not working in linux?
	fmt.Fprint(tm.Screen, "\033[?25h")
}

type flushWriter struct {
	io.Writer
	ui *UI
}

// TODO: wrap lines
func (f *flushWriter) Write(p []byte) (int, error) {
	n, err := f.Writer.Write(p)
	f.ui.draw()
	return n, err
}

func formatHeader(command []string) string {
	return "filewatcher │ " + strings.Join(command, " ")
}
