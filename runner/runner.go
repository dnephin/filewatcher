package runner

import (
	"context"
	"os"
	"os/exec"
	"path"
	"path/filepath"
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
	chUpdate chan Update
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
		chUpdate: make(chan Update, 1),
	}, func() { close(events) }
}

func (runner *Runner) start(ctx context.Context) {
	go runner.runLoop(ctx)
	go scanInput(runner.chUpdate)
}

func (runner *Runner) runLoop(ctx context.Context) {
	// TODO: store as map to override keys
	environ := os.Environ()
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-runner.events:
			// FIXME: I'm not sure how this empty event gets created
			if event.Name == "" && event.Op == 0 {
				return
			}
			runner.run(event, environ)
		case update := <-runner.chUpdate:
			if update.Env != "" {
				environ = append(environ, update.Env)
			}
		}
	}
}

// HandleEvent checks runs the command if the event was a Write event
func (runner *Runner) HandleEvent(event fsnotify.Event) {
	if !runner.shouldHandle(event) {
		return
	}

	// Send the event to an unbuffered channel so that on events floods only
	// one event is run, and the rest are dropped.
	select {
	case runner.events <- event:
	default:
		log.Debugf("Events queued, skipping: %s", event.Name)
	}
}

func (runner *Runner) run(event fsnotify.Event, environ []string) {
	start := time.Now()
	command := runner.buildCommand(event.Name)
	ui.PrintStart(command)

	err := run(command, event.Name, environ)
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
		case "relative_dir":
			return addDotSlash(filepath.Dir(filename))
		}
		return key
	}

	output := []string{}
	for _, arg := range runner.command {
		output = append(output, os.Expand(arg, mapping))
	}
	return output
}

func run(command []string, filename string, environ []string) error {
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(environ,
		"TEST_DIRECTORY="+addDotSlash(filepath.Dir(filename)),
		"TEST_FILENAME="+addDotSlash(filename))
	return cmd.Run()
}

func addDotSlash(filename string) string {
	return "." + string(filepath.Separator) + filename
}
