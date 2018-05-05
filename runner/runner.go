package runner

import (
	"os"
	"os/exec"
	"path"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/dnephin/filewatcher/files"
	"github.com/dnephin/filewatcher/ui"
	"github.com/fsnotify/fsnotify"
)

// Runner executes commands when an included file is modified
type Runner struct {
	excludes *files.ExcludeList
	command  []string
	events   chan fsnotify.Event
	eventOp  fsnotify.Op
}

// NewRunner creates a new Runner
func NewRunner(
	excludes *files.ExcludeList,
	eventOp fsnotify.Op,
	command []string,
) (*Runner, func()) {
	events := make(chan fsnotify.Event)
	return &Runner{
		excludes: excludes,
		command:  command,
		events:   events,
		eventOp:  eventOp,
	}, func() { close(events) }
}

func (runner *Runner) start() {
	for {
		select {
		case event := <-runner.events:
			runner.handle(event)
		}
	}
}

// HandleEvent checks runs the command if the event was a Write event
func (runner *Runner) HandleEvent(event fsnotify.Event) {
	if !runner.shouldHandle(event) {
		return
	}

	// Handle events in another goroutine so that on events floods only
	// one event is run, and the rest are dropped.
	select {
	case runner.events <- event:
	default:
		log.Debugf("Events queued, skipping: %s", event.Name)
	}
}

func (runner *Runner) handle(event fsnotify.Event) {
	start := time.Now()
	command := runner.buildCommand(event.Name)
	ui.PrintStart(command)

	err := run(command)
	ui.PrintEnd(time.Since(start), event.Name, err)
}

func (runner *Runner) shouldHandle(event fsnotify.Event) bool {
	if event.Op&runner.eventOp == 0 {
		log.Debugf("Skipping excluded event: %s (%v)", event.Op, event.Op&runner.eventOp)
		return false
	}

	filename := event.Name
	if runner.excludes.IsMatch(filename) {
		log.Debugf("Skipping excluded file: %s", filename)
		return false
	}

	return true
}

func (runner *Runner) buildCommand(filename string) []string {
	mapping := func(key string) string {
		switch key {
		case "filepath":
			return filename
		case "dir":
			return path.Dir(filename)
		}
		return key
	}

	output := []string{}
	for _, arg := range runner.command {
		output = append(output, os.Expand(arg, mapping))
	}
	return output
}

func run(command []string) error {
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
