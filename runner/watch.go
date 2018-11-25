package runner

import (
	"context"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/dnephin/filewatcher/files"
	"github.com/fsnotify/fsnotify"
)

// WatchOptions passed to watch
type WatchOptions struct {
	IdleTimeout time.Duration
	Runner      *Runner
}

// Watch for events from the watcher and handle them with the runner
func Watch(watcher *fsnotify.Watcher, opts WatchOptions) error {
	runner := opts.Runner
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go runner.start(ctx)

	for {
		select {
		case <-time.After(opts.IdleTimeout):
			log.Warnf("Idle timeout hit: %s", opts.IdleTimeout)
			return nil

		case event := <-watcher.Events:
			log.Debugf("Event: %s", event)

			if isNewDir(event, runner.excludes) {
				log.Debugf("Watching new directory: %s", event.Name)
				watcher.Add(event.Name)
				continue
			}
			runner.HandleEvent(event)

		case err := <-watcher.Errors:
			return err
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
