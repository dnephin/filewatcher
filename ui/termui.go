package ui

import (
	"fmt"
	"io"
	"strings"
	"time"

	tm "github.com/buger/goterm"
	termbox "github.com/nsf/termbox-go"
)

const border = "─ │ ┌ ┐ └ ┘"

// UI stores terminal UI elements
type UI struct {
	header *tm.Box
	footer *tm.Box
	logs   *tm.Box
}

// NewUI elements
func NewUI(command []string) (*UI, error) {
	err := termbox.Init()
	if err != nil {
		return nil, err
	}

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

	ui.Header(command)
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
	termbox.Close()
}

func (ui *UI) Handle(chEvents chan Event) error {
	for {
		select {
		case event := <- chEvents:
			switch event.Type {
			case EventRunFinished:
				ui.Footer(event.Message)
			case EventClear:
				ui.logs.Buf.Reset()
				ui.draw()
			}
		}
	}
}

type flushWriter struct {
	io.Writer
	ui *UI
}

// TODO: wrap lines
// TODO: either lock or send events to Handle() to deal with concurrency
func (f *flushWriter) Write(p []byte) (int, error) {
	n, err := f.Writer.Write(p)
	f.ui.draw()
	return n, err
}

func formatHeader(command []string) string {
	return "filewatcher │ " + strings.Join(command, " ")
}

// EventType is an enumeration of events that can be triggered by a user
type Event struct {
	Type EventType
	Err error
	Message string
}

type EventType int

const (
	EventClear EventType = iota
	EventReset
	EventUpdateCommand
	EventRunFinished
)

const (
	KeyC termbox.Key = 67
	KeyR termbox.Key = 82
	KeyU termbox.Key = 85
)

func RunKeyPoller(chEvents chan Event) error {
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyCtrlC:
				return nil
			case KeyC:
				chEvents <- NewEvent(EventClear)
			case KeyR:
				chEvents <- NewEvent(EventReset)
			case KeyU:
				chEvents <- NewEvent(EventUpdateCommand)
			}
		case termbox.EventError:
			return ev.Err
		}
	}
}

func NewEvent(eventType EventType) Event {
	return Event{Type: eventType}
}

func NewRunFinishedEvent(err error, filename string, elapsed time.Duration) Event {
	msg := "OK"
	if err != nil {
		msg = err.Error()
	}
	return Event{
		Err: err,
		Type: EventRunFinished,
		Message: fmt.Sprintf("%s │ %s | %s", msg, filename, elapsed),
	}
}
