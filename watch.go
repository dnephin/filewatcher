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
)

type options struct {
	verbose bool
	quiet   bool
	exclude []string
	dirs    []string
	depth   int
	command []string
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
	setupLogging(opts)

	opts.command = flag.Args()
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

	watcher, err := buildWatcher(files.WalkDirectories(opts.dirs, opts.depth, excludeList))
	if err != nil {
		log.Fatalf("Error setting up watcher: %s", err)
	}
	defer watcher.Close()

	handler := runner.NewRunner(excludeList, opts.command)
	if err = runner.Watch(watcher, handler); err != nil {
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
