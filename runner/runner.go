package runner

import (
	"io"
	"os"
	"os/exec"
	"path"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/dnephin/filewatcher/files"
	fsnotify "gopkg.in/fsnotify.v1"
)

// Streams used by processes created by Runner
type Streams struct {
	Out io.Writer
	Err io.Writer
	In  io.Reader
}

// Runner executes commands when an included file is modified
type Runner struct {
	excludes *files.ExcludeList
	command  []string
	streams  Streams
}

// NewRunner creates a new Runner
func NewRunner(excludes *files.ExcludeList, command []string, streams Streams) *Runner {
	return &Runner{
		excludes: excludes,
		command:  command,
		streams:  streams,
	}
}

// Excludes returns the exclude list
func (runner *Runner) Excludes() *files.ExcludeList {
	return runner.excludes
}

// HandleEvent checks runs the command if the event was a Write event
func (runner *Runner) HandleEvent(event fsnotify.Event) (bool, error) {
	if event.Op&fsnotify.Write != fsnotify.Write {
		return false, nil
	}

	filename := event.Name
	if runner.excludes.IsMatch(filename) {
		log.Debugf("Skipping excluded file: %s", filename)
		return false, nil
	}

	return true, runner.Run(filename)
}

// Run the command for the given filename
func (runner *Runner) Run(filename string) error {
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
	log.Infof("Running: %s", strings.Join(output, " "))

	cmd := exec.Command(output[0], output[1:]...)
	cmd.Stdout = runner.streams.Out
	cmd.Stderr = runner.streams.Err
	cmd.Stdin = runner.streams.In
	return cmd.Run()
}
