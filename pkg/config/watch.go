package config

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// ChangeHandler is called when a watched config file is reloaded successfully.
type ChangeHandler func(cfg *Config)

// Watcher polls a config file for changes and reloads it.
type Watcher struct {
	path     string
	interval time.Duration
	onChange ChangeHandler
	onError  func(error)

	mu       sync.Mutex
	current  *Config
	modTime  time.Time
	stopCh   chan struct{}
	stopped  bool
	stopOnce sync.Once
}

// WatchFile starts polling path for modifications.
// interval defaults to 1s when <= 0.
func WatchFile(path string, interval time.Duration, onChange ChangeHandler) (*Watcher, error) {
	if onChange == nil {
		return nil, fmt.Errorf("onChange must not be nil")
	}
	if interval <= 0 {
		interval = time.Second
	}

	cfg, err := LoadFromFile(path)
	if err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	w := &Watcher{
		path:     path,
		interval: interval,
		onChange: onChange,
		current:  cfg,
		modTime:  info.ModTime(),
		stopCh:   make(chan struct{}),
	}

	// Deliver initial config.
	onChange(cfg.Clone())

	go w.loop()
	return w, nil
}

// OnError sets an optional error callback for reload failures.
func (w *Watcher) OnError(fn func(error)) {
	w.mu.Lock()
	w.onError = fn
	w.mu.Unlock()
}

// Current returns a copy of the latest successfully loaded config.
func (w *Watcher) Current() *Config {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.current == nil {
		return nil
	}
	return w.current.Clone()
}

// Stop ends the watch loop.
func (w *Watcher) Stop() {
	w.stopOnce.Do(func() {
		w.mu.Lock()
		w.stopped = true
		w.mu.Unlock()
		close(w.stopCh)
	})
}

func (w *Watcher) loop() {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopCh:
			return
		case <-ticker.C:
			w.poll()
		}
	}
}

func (w *Watcher) poll() {
	info, err := os.Stat(w.path)
	if err != nil {
		w.reportError(fmt.Errorf("stat config: %w", err))
		return
	}

	w.mu.Lock()
	prev := w.modTime
	w.mu.Unlock()

	if !info.ModTime().After(prev) {
		return
	}

	cfg, err := LoadFromFile(w.path)
	if err != nil {
		w.reportError(fmt.Errorf("reload config: %w", err))
		return
	}
	if err := cfg.Validate(); err != nil {
		w.reportError(fmt.Errorf("invalid config: %w", err))
		return
	}

	w.mu.Lock()
	w.modTime = info.ModTime()
	w.current = cfg
	onChange := w.onChange
	w.mu.Unlock()

	onChange(cfg.Clone())
}

func (w *Watcher) reportError(err error) {
	w.mu.Lock()
	fn := w.onError
	w.mu.Unlock()
	if fn != nil {
		fn(err)
	}
}

// Clone returns a deep copy of the config.
func (c *Config) Clone() *Config {
	if c == nil {
		return nil
	}
	out := *c
	if c.ALPN != nil {
		out.ALPN = append([]string(nil), c.ALPN...)
	}
	return &out
}
