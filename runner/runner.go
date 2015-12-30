package runner

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/fsnotify.v1"
)

// Runner executes commands when an included file is modified
type Runner struct {
	excludes []string
	command  []string
}

// NewRunner creates a new Runner
func NewRunner(excludes, command []string) (*Runner, error) {
	for _, exclude := range excludes {
		if _, err := filepath.Match(exclude, "."); err != nil {
			return nil, err
		}
	}

	if len(command) == 0 {
		return nil, fmt.Errorf("A command is required.")
	}

	runner := Runner{
		excludes: excludes,
		command:  command,
	}
	return &runner, nil
}

// HandleEvent checks runs the command if the event was a Write event
func (runner *Runner) HandleEvent(event fsnotify.Event) error {
	if event.Op&fsnotify.Write != fsnotify.Write {
		return nil
	}

	filename := event.Name
	if runner.isExcluded(filename) {
		log.Debugf("Skipping excluded file: %s", filename)
		return nil
	}

	return runner.Run(filename)
}

func (runner *Runner) isExcluded(filename string) bool {
	for _, exclude := range runner.excludes {
		// exclude patterns were already validated in NewRunner, so error
		// can be ignored here
		match, _ := filepath.Match(exclude, filename)
		if match {
			return true
		}
	}
	return false
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
	return cmd.Run()
}
