package throttle

import (
	"sync"
	"time"

	"github.com/LinharesAron/jotunn/internal/logger"
)

type StandardThrottler struct {
	Throttler
	mu   sync.Mutex
	cond *sync.Cond

	startTime time.Time

	blocked bool

	lastRequest time.Time
	lastTrigger time.Time

	reqCount  int
	threshold int

	backoff                   time.Duration
	recoveredSinceLastTrigger bool
}

func (s *StandardThrottler) Wait() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for s.blocked {
		s.cond.Wait()
	}

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
	s.reqCount++
}

func (s *StandardThrottler) MarkRecovered() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.recoveredSinceLastTrigger = true
}

func (s *StandardThrottler) Trigger() {
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
		s.recoveredSinceLastTrigger = false
	} else {
		s.backoff *= 2
		if s.backoff > 50*time.Minute {
			s.backoff = 50 * time.Minute
		}

		rpm = max(int(float64(rpm)*0.9), 10)
		logger.Warn("[StandardThrottler] Threshold reduced by 10%% → %d RPM", rpm)
		if rpm < s.threshold {
			s.threshold = rpm
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

	s.cond.Broadcast()
	s.mu.Unlock()

	now = time.Now().Format("15:04:05")
	logger.Warn("[StandardThrottler] [%s] Cooldown complete, resuming operations", now)
}
