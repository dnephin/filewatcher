package main

import (
	"fmt"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/dnephin/filewatcher/files"
	"github.com/dnephin/filewatcher/runner"
	flag "github.com/spf13/pflag"
	"gopkg.in/fsnotify.v1"
	"github.com/pkg/errors"
)

type options struct {
	verbose bool
	quiet   bool
	exclude []string
	dirs    []string
	depth   int
	command []string
}

func watch(watcher *fsnotify.Watcher, runner *runner.Runner, shutdown chan struct{}) error {
	for {
		select {
		case event := <-watcher.Events:
			log.Debugf("Event: %s", event)

			if isNewDir(event, runner.Excludes()) {
				log.Debugf("Watching new directory: %s", event.Name)
				watcher.Add(event.Name)
				continue
			}

			err := runner.HandleEvent(event)
			if err != nil {
				log.Warnf("Error while handling %s: %s", event, err)
			}
		case err := <-watcher.Errors:
			return err
		case <- shutdown:
			return nil
		}
	}
}

func isNewDir(event fsnotify.Event, exclude *files.ExcludeList) bool {
	if event.Op&fsnotify.Create != fsnotify.Create {
		return false
	}

	fileInfo, err := os.Stat(event.Name)
	if err != nil {
		log.Warnf("Failed to stat %s: %s", event.Name, err)
		return false
	}

	return fileInfo.IsDir() && !exclude.IsMatch(event.Name)
}

func buildWatcher(dirs []string) (*fsnotify.Watcher, error) {
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

func setupFlags() *options {
	opts := options{}
	flag.BoolVarP(&opts.verbose, "verbose", "v", false, "Verbose")
	flag.BoolVarP(&opts.quiet, "quiet", "q", false, "Quiet")
	flag.StringSliceVarP(&opts.exclude, "exclude", "x", nil, "Exclude file patterns")
	flag.StringSliceVarP(&opts.dirs, "directory", "d", []string{"."}, "Directories to watch")
	flag.IntVarP(&opts.depth, "depth", "L", 5, "Descend only level directories deep")
	return &opts
}

func main() {
	opts := setupFlags()
	cmd := flag.CommandLine
	cmd.Init(os.Args[0], flag.ExitOnError)
	cmd.SetInterspersed(false)
	flag.Usage = func() {
		out := os.Stderr
		fmt.Fprintf(out, "Usage:\n  %s [OPTIONS] COMMAND ARGS... \n\n", os.Args[0])
		fmt.Fprint(out, "Options:\n")
		cmd.PrintDefaults()
	}
	flag.Parse()
	opts.command = flag.Args()
	if len(opts.command) == 0 {
		log.Fatalf("A command argument is required")
	}

	setupLogging(opts)
	shutdown := make(chan struct{}, 1)
	if err := run(opts, shutdown); err != nil {
		log.Fatal(err.Error())
	}
}

func run(opts *options, shutdown chan struct{}) error {
	excludeList, err := files.NewExcludeList(opts.exclude)
	if err != nil {
		return errors.Wrap(err, "failed to create exclude list")
	}

	watcher, err := buildWatcher(files.WalkDirectories(opts.dirs, opts.depth, excludeList))
	if err != nil {
		return errors.Wrap(err, "failed to setup watcher")
	}
	defer watcher.Close()

	runner, err := runner.NewRunner(excludeList, opts.command)
	if err != nil {
		return errors.Wrap(err, "failed to setup runner")
	}

	err = watch(watcher, runner, shutdown)
	return errors.Wrap(err, "error while filewatching")
}

func setupLogging(opts *options) {
	if opts.verbose {
		log.SetLevel(log.DebugLevel)
	}
	if opts.quiet {
		log.SetLevel(log.WarnLevel)
	}
}
