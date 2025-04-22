package core

import (
	"fmt"
	"sync"

	"github.com/LinharesAron/jotunn/internal/types"
)

type RetryTracker struct {
	mu     sync.Mutex
	counts map[string]int
	limit  int
}

func NewRetryTracker(limit int) *RetryTracker {
	return &RetryTracker{
		counts: make(map[string]int),
		limit:  limit,
	}
}

func (r *RetryTracker) ShouldRetry(attempt types.Attempt) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := fmt.Sprintf("%s:%s", attempt.Username, attempt.Password)
	r.counts[key]++

	if r.counts[key] > r.limit {
		delete(r.counts, key)
		return false
	}
	return true
}
