package beads

import (
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher watches for changes to beads files
type Watcher struct {
	projectPath string
	watcher     *fsnotify.Watcher
	onChange    func()
	debounce    time.Duration
	stopCh      chan struct{}
	mu          sync.Mutex
	lastEvent   time.Time
}

// NewWatcher creates a new beads file watcher
func NewWatcher(projectPath string, onChange func()) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &Watcher{
		projectPath: projectPath,
		watcher:     fsWatcher,
		onChange:    onChange,
		debounce:    100 * time.Millisecond, // Debounce rapid changes
		stopCh:      make(chan struct{}),
	}

	return w, nil
}

// Start begins watching for beads file changes
func (w *Watcher) Start() error {
	reader := NewReader(w.projectPath)
	if !reader.HasBeads() {
		return nil // No beads directory to watch
	}

	// Watch the .beads directory
	if err := w.watcher.Add(reader.BeadsPath()); err != nil {
		return err
	}

	go w.watchLoop()
	return nil
}

// Stop stops the watcher
func (w *Watcher) Stop() error {
	close(w.stopCh)
	return w.watcher.Close()
}

// watchLoop handles file system events
func (w *Watcher) watchLoop() {
	var debounceTimer *time.Timer

	for {
		select {
		case <-w.stopCh:
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			return

		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}

			// Only care about writes to the issues file
			if event.Op&fsnotify.Write != fsnotify.Write {
				continue
			}

			// Debounce: wait for activity to settle
			w.mu.Lock()
			w.lastEvent = time.Now()
			w.mu.Unlock()

			if debounceTimer != nil {
				debounceTimer.Stop()
			}

			debounceTimer = time.AfterFunc(w.debounce, func() {
				w.mu.Lock()
				elapsed := time.Since(w.lastEvent)
				w.mu.Unlock()

				if elapsed >= w.debounce {
					w.onChange()
				}
			})

		case _, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			// Log errors but continue watching
		}
	}
}

// SetDebounce sets the debounce duration
func (w *Watcher) SetDebounce(d time.Duration) {
	w.debounce = d
}
