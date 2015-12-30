package main

import (
	"fmt"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/dnephin/filewatcher/runner"
	flag "github.com/spf13/pflag"
	"gopkg.in/fsnotify.v1"
)

var (
	verbose = flag.BoolP("verbose", "v", false, "Verbose")
	quiet   = flag.BoolP("quiet", "q", false, "Quiet")
	exclude = flag.StringSliceP("exclude", "x", nil, "Exclude file patterns")
	dirs    = flag.StringSliceP("directory", "d", []string{"."}, "Directories to watch")
)

func watch(watcher *fsnotify.Watcher, runner *runner.Runner) error {
	for {
		select {
		case event := <-watcher.Events:
			log.Debugf("Event: %s", event)
			err := runner.HandleEvent(event)
			if err != nil {
				log.Warnf("Error while handling %s: %s", event, err)
			}
		case err := <-watcher.Errors:
			return err
		}
	}
}

func buildWatcher(dirs, exclude []string) (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	log.Infof("Watching directories: %s", strings.Join(dirs, ", "))
	for _, dir := range dirs {
		log.Debugf("Adding new watch: %s", dir)
		if err = watcher.Add(dir); err != nil {
			return nil, err
		}
	}
	return watcher, nil
}

func main() {
	cmd := flag.CommandLine
	cmd.Init(os.Args[0], flag.ExitOnError)
	flag.Usage = func() {
		out := os.Stderr
		fmt.Fprintf(out, "Usage:\n  %s [OPTIONS] COMMAND ARG1 ARGN... \n\n", os.Args[0])
		fmt.Fprintf(out, "Options:\n")
		cmd.PrintDefaults()
	}
	flag.Parse()

	if *verbose {
		log.SetLevel(log.DebugLevel)
	}
	if *quiet {
		log.SetLevel(log.WarnLevel)
	}

	watcher, err := buildWatcher(*dirs, *exclude)
	if err != nil {
		log.Fatalf("Error setting up watcher: %s", err)
	}
	defer watcher.Close()

	runner, err := runner.NewRunner(*exclude, flag.Args())
	if err != nil {
		log.Fatalf("Error setting up runner: %s", err)
	}

	if err = watch(watcher, runner); err != nil {
		log.Fatalf("Error during watch: %s", err)
	}
}
