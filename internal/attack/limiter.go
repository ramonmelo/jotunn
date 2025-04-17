package attack

import (
	"sync"
	"time"

	"github.com/LinharesAron/jotunn/internal/logger"
)

type RateLimitManager struct {
	mu          sync.Mutex
	cond        *sync.Cond
	blocked     bool
	lastRequest time.Time

	lastTrigger               time.Time
	startTime                 time.Time
	reqCount                  int
	threshold                 int
	backoff                   time.Duration
	recoveredSinceLastTrigger bool
}

func NewRateLimitManager(threshold int) *RateLimitManager {
	mgr := &RateLimitManager{
		threshold:                 threshold,
		backoff:                   5 * time.Minute,
		recoveredSinceLastTrigger: true,
	}
	mgr.cond = sync.NewCond(&mgr.mu)
	mgr.startTime = time.Now()
	return mgr
}

func (r *RateLimitManager) RegisterRequest() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.reqCount++
}

func (r *RateLimitManager) WaitIfBlocked() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for r.blocked {
		r.cond.Wait()
	}
}

func (r *RateLimitManager) Trigger() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.blocked {
		return
	}

	r.blocked = true
	r.lastTrigger = time.Now()

	if r.recoveredSinceLastTrigger {
		r.recoveredSinceLastTrigger = false
	} else {
		r.backoff *= 2
		if r.backoff > 50*time.Minute {
			r.backoff = 50 * time.Minute
		}
	}

	elapsed := time.Since(r.startTime).Minutes()
	rpm := int(float64(r.reqCount) / elapsed)

	logger.Error("[RateLimit] Triggered in the %d try - estimated RPM: %d\n", r.reqCount, rpm)

	if rpm < r.threshold {
		r.threshold = rpm
	}

	go r.cooldown()
}

func (r *RateLimitManager) WaitCadence() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.threshold <= 0 {
		return
	}

	now := time.Now()
	minInterval := time.Minute / time.Duration(r.threshold)
	elapsed := now.Sub(r.lastRequest)

	if elapsed < minInterval {
		time.Sleep(minInterval - elapsed)
	}

	r.lastRequest = time.Now()
}

func (r *RateLimitManager) cooldown() {
	r.mu.Lock()
	duration := r.backoff
	r.mu.Unlock()

	logger.Error("[RateLimit] Cooling down for %s...\n", duration)
	time.Sleep(duration)

	r.mu.Lock()
	r.blocked = false
	r.reqCount = 0
	r.startTime = time.Now()
	r.mu.Unlock()

	logger.Error("[RateLimit] Cooldown complete, resuming operations")
	r.cond.Broadcast()
}

func (r *RateLimitManager) MarkRecovered() {
	r.mu.Lock()
	r.recoveredSinceLastTrigger = true
	r.mu.Unlock()
}
