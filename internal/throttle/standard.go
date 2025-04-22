package throttle

import (
	"slices"
	"sync"
	"time"

	"github.com/LinharesAron/jotunn/internal/core"
	"github.com/LinharesAron/jotunn/internal/logger"
	"github.com/LinharesAron/jotunn/internal/types"
)

type StandardThrottler struct {
	mu   sync.Mutex
	cond *sync.Cond

	startTime time.Time

	blocked bool

	lastRequest time.Time
	lastTrigger time.Time

	reqCount  int
	threshold int

	backoff                   time.Duration
	throttleCodes             []int
	recoveredSinceLastTrigger bool
}

func (s *StandardThrottler) RegisterRequest() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.reqCount++
}

func (s *StandardThrottler) WaitIfBlocked() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for s.blocked {
		s.cond.Wait()
	}
}

func (s *StandardThrottler) WaitCadence() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.threshold <= 0 {
		return
	}

	now := time.Now()
	minInterval := time.Minute / time.Duration(s.threshold)
	elapsed := now.Sub(s.lastRequest)

	if elapsed < minInterval {
		time.Sleep(minInterval - elapsed)
	}

	s.lastRequest = time.Now()
}

func (s *StandardThrottler) IsThrottling(statusCode int) bool {
	return slices.Contains(s.throttleCodes, statusCode)
}

func (s *StandardThrottler) HandleThrottle(statusCode int, dispatcher *core.Dispatcher, attempt types.Attempt) bool {
	if !s.IsThrottling(statusCode) {
		return false
	}

	if err := dispatcher.Retry(attempt); err != nil {
		logger.Warn("[StandardThrottler] Retry limit reached for %s:%s – ignoring attempt → %s", attempt.Username, attempt.Password, err)
	}

	s.trigger()
	return true
}

func (s *StandardThrottler) MarkRecovered() {
	s.mu.Lock()
	s.recoveredSinceLastTrigger = true
	s.mu.Unlock()
}

func (s *StandardThrottler) trigger() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.blocked {
		return
	}

	s.blocked = true
	s.lastTrigger = time.Now()

	elapsed := time.Since(s.startTime).Minutes()
	rpm := int(float64(s.reqCount) / elapsed)
	logger.Warn("[StandardThrottler] Triggered after %d attempts – estimated RPM: %d", s.reqCount, rpm)

	if s.recoveredSinceLastTrigger {
		rpm = max(int(float64(rpm)*0.9), 10)
		logger.Warn("[StandardThrottler] Threshold reduced by 10%% → %d RPM", rpm)

		if rpm < s.threshold {
			s.threshold = rpm
		}

		s.recoveredSinceLastTrigger = false
	} else {
		s.backoff *= 2
		if s.backoff > 50*time.Minute {
			s.backoff = 50 * time.Minute
		}
	}

	go s.cooldown()
}

func (s *StandardThrottler) cooldown() {
	s.mu.Lock()
	duration := s.backoff
	s.mu.Unlock()

	now := time.Now().Format("15:04:05")
	logger.Warn("[StandardThrottler] [%s] Cooling down for %s...", now, duration)
	time.Sleep(duration)

	s.mu.Lock()
	s.blocked = false
	s.reqCount = 0
	s.startTime = time.Now()
	s.mu.Unlock()

	now = time.Now().Format("15:04:05")
	logger.Warn("[StandardThrottler] [%s] Cooldown complete, resuming operations\n", now)
	s.cond.Broadcast()
}
