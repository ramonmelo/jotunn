package attack

import (
	"slices"
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
	rateLimitCodes            []int
}

func NewRateLimitManager(threshold int, rateLimitCodes []int) *RateLimitManager {
	mgr := &RateLimitManager{
		threshold:                 threshold,
		backoff:                   5 * time.Minute,
		recoveredSinceLastTrigger: true,
		rateLimitCodes:            rateLimitCodes,
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

func (r *RateLimitManager) IsRateLimited(statusCode int) bool {
	return slices.Contains(r.rateLimitCodes, statusCode)
}

func (r *RateLimitManager) HandleIfRateLimited(statusCode int, retries chan<- Attempt, attempt Attempt) bool {
	if r.IsRateLimited(statusCode) {
		r.TriggerAndRetry()
		retries <- attempt
		return true
	}
	return false
}

func (r *RateLimitManager) TriggerAndRetry() {
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

	logger.Error("[RateLimit] Triggered after %d attempts – estimated RPM: %d", r.reqCount, rpm)

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

	now := time.Now().Format("15:04:05")
	logger.Error("[RateLimit] [%s] Cooling down for %s...", now, duration)
	time.Sleep(duration)

	r.mu.Lock()
	r.blocked = false
	r.reqCount = 0
	r.startTime = time.Now()
	r.mu.Unlock()

	now = time.Now().Format("15:04:05")
	logger.Error("[RateLimit] [%s] Cooldown complete, resuming operations\n", now)
	r.cond.Broadcast()
}

func (r *RateLimitManager) MarkRecovered() {
	r.mu.Lock()
	r.recoveredSinceLastTrigger = true
	r.mu.Unlock()
}
