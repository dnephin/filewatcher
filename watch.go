package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/dnephin/filewatcher/files"
	"github.com/dnephin/filewatcher/runner"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/pflag"
)

type options struct {
	verbose     bool
	quiet       bool
	exclude     []string
	dirs        []string
	depth       int
	command     []string
	idleTimeout time.Duration
	events      eventOpt
}

type eventOpt struct {
	value fsnotify.Op
}

func (o *eventOpt) Set(value string) error {
	var op fsnotify.Op
	switch value {
	case "create":
		op = fsnotify.Create
	case "write":
		op = fsnotify.Write
	case "remove":
		op = fsnotify.Remove
	case "rename":
		op = fsnotify.Rename
	case "chmod":
		op = fsnotify.Chmod
	default:
		return fmt.Errorf("unknown event: %s", value)
	}
	o.value = o.value | op
	return nil
}

func (o *eventOpt) Type() string {
	return "event"
}

func (o *eventOpt) String() string {
	return fmt.Sprintf("%s", o.value)
}

func (o *eventOpt) Value() fsnotify.Op {
	if o.value == 0 {
		return fsnotify.Write | fsnotify.Create
	}
	return o.value
}

func setupFlags(args []string) *options {
	flags := pflag.NewFlagSet(args[0], pflag.ContinueOnError)
	opts := options{}
	flags.BoolVarP(&opts.verbose, "verbose", "v", false, "Verbose")
	flags.BoolVarP(&opts.quiet, "quiet", "q", false, "Quiet")
	flags.StringSliceVarP(&opts.exclude, "exclude", "x", nil, "Exclude file patterns")
	flags.StringSliceVarP(&opts.dirs, "directory", "d", []string{"."}, "Directories to watch")
	flags.IntVarP(&opts.depth, "depth", "L", 5, "Descend only level directories deep")
	flags.DurationVar(&opts.idleTimeout, "idle-timeout", 10*time.Minute,
		"Exit after idle timeout")
	flags.VarP(&opts.events, "event", "e",
		"events to watch (create,write,remove,rename,chmod)")

	flags.SetInterspersed(false)
	flags.Usage = func() {
		out := os.Stderr
		fmt.Fprintf(out, "Usage:\n  %s [OPTIONS] COMMAND ARGS... \n\n", os.Args[0])
		fmt.Fprint(out, "Options:\n")
		flags.PrintDefaults()
	}
	if err := flags.Parse(args[1:]); err != nil {
		os.Exit(2)
	}
	opts.command = flags.Args()
	return &opts
}

func main() {
	opts := setupFlags(os.Args)
	setupLogging(opts)

	if len(opts.command) == 0 {
		log.Fatalf("A command argument is required.")
	}
	run(opts)
}

func run(opts *options) {
	excludeList, err := files.NewExcludeList(opts.exclude)
	if err != nil {
		log.Fatalf("Error creating exclude list: %s", err)
	}

	dirs := files.WalkDirectories(opts.dirs, opts.depth, excludeList)
	watcher, err := buildWatcher(dirs)
	if err != nil {
		log.Fatalf("Error setting up watcher: %s", err)
	}
	defer watcher.Close()

	log.Debugf("Handling events: %s", opts.events.Value())
	handler, cleanup := runner.NewRunner(excludeList, opts.events.Value(), opts.command)
	defer cleanup()
	watchOpts := runner.WatchOptions{
		IdleTimeout: opts.idleTimeout,
		Runner:      handler,
	}
	if err = runner.Watch(watcher, watchOpts); err != nil {
		log.Fatalf("Error during watch: %s", err)
	}
}

func setupLogging(opts *options) {
	if opts.verbose {
		log.SetLevel(log.DebugLevel)
	}
	if opts.quiet {
		log.SetLevel(log.WarnLevel)
	}
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
