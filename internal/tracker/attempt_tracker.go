package tracker

import (
	"crypto/sha1"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type AttemptTracker struct {
	mu     sync.RWMutex
	seen   map[string]struct{}
	saveCh chan string
	wg     sync.WaitGroup
}

func newAttemptTracker(basePath, filename string) Tracker {
	path := filepath.Join(basePath, filename)
	attempts := &AttemptTracker{
		seen:   make(map[string]struct{}),
		saveCh: make(chan string, 100),
	}
	attempts.loadFromFile(path)
	StartSaver(path, attempts.saveCh, &attempts.wg)
	return attempts
}

func (t *AttemptTracker) HasSeen(keys ...string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	id := attemptID(keys...)
	_, ok := t.seen[id]
	return ok
}

func (t *AttemptTracker) Mark(keys ...string) {
	t.mu.Lock()
	id := attemptID(keys...)
	t.seen[id] = struct{}{}
	t.mu.Unlock()
	t.saveCh <- id
}

func (t *AttemptTracker) loadFromFile(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if line != "" {
			t.seen[line] = struct{}{}
		}
	}
}

func (t *AttemptTracker) Close() {
	close(t.saveCh)
	t.wg.Wait()
}

func attemptID(keys ...string) string {
	h := sha1.Sum([]byte(strings.Join(keys, ":")))
	return hex.EncodeToString(h[:])
}
