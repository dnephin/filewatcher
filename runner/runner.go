package runner

import (
	"os"
	"os/exec"
	"path"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/dnephin/filewatcher/files"
	"gopkg.in/fsnotify.v1"
)

// Runner executes commands when an included file is modified
type Runner struct {
	excludes *files.ExcludeList
	command  []string
}

// NewRunner creates a new Runner
func NewRunner(excludes *files.ExcludeList, command []string) (*Runner, error) {
	runner := Runner{
		excludes: excludes,
		command:  command,
	}
	return &runner, nil
}

// Excludes returns the exclude list
func (runner *Runner) Excludes() *files.ExcludeList {
	return runner.excludes
}

// HandleEvent checks runs the command if the event was a Write event
func (runner *Runner) HandleEvent(event fsnotify.Event) error {
	if event.Op&fsnotify.Write != fsnotify.Write {
		return nil
	}

	filename := event.Name
	if runner.excludes.IsMatch(filename) {
		log.Debugf("Skipping excluded file: %s", filename)
		return nil
	}

	return runner.Run(filename)
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
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
