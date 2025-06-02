package tracker

import (
	"path/filepath"
	"strings"
	"sync"
)

type CredentialTracker struct {
	mu     sync.RWMutex
	found  map[string]struct{}
	saveCh chan string
	wg     sync.WaitGroup
}

func newCredentialTracker(basePath string) Tracker {
	path := filepath.Join(basePath, ".found_credentials")
	credentials := &CredentialTracker{
		found:  make(map[string]struct{}),
		saveCh: make(chan string, 100),
	}
	StartSaver(path, credentials.saveCh, &credentials.wg)
	return credentials
}

func (t *CredentialTracker) HasSeen(keys ...string) bool {
	if len(keys) != 2 {
		return false
	}

	t.mu.RLock()
	defer t.mu.RUnlock()

	id := strings.Join(keys, ":")
	_, ok := t.found[id]
	return ok
}

func (t *CredentialTracker) Mark(keys ...string) {
	if len(keys) != 2 {
		return
	}

	t.mu.Lock()
	id := strings.Join(keys, ":")
	t.found[id] = struct{}{}
	t.mu.Unlock()
	t.saveCh <- id
}

func (t *CredentialTracker) Close() {
	close(t.saveCh)
	t.wg.Wait()
}
