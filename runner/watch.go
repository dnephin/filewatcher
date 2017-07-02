package runner

import (
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/dnephin/filewatcher/files"
	"gopkg.in/fsnotify.v1"
)

// WatchOptions passed to watch
type WatchOptions struct {
	IdleTimeout time.Duration
}

// Watch for events from the watcher and handle them with the runner
func Watch(watcher *fsnotify.Watcher, runner *Runner, opts WatchOptions) error {
	events := make(chan fsnotify.Event)
	defer close(events)
	go handleEvents(runner, events)

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

			// Handle events in another goroutine so that on events floods only
			// one event is run, and the rest are dropped.
			select {
			case events <- event:
			default:
				log.Debugf("Events queued, skipping: %s", event.Name)
			}

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

func handleEvents(runner *Runner, events chan fsnotify.Event) {
	for {
		select {
		case event := <-events:
			runner.HandleEvent(event)
		}
	}
}
