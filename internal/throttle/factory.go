package throttle

import (
	"sync"
	"time"

	"github.com/LinharesAron/jotunn/internal/config"
)

func New(cfg *config.AttackConfig) Throttler {
	if cfg.UseNoLimit {
		return &NoLimitThrottle{}
	}

	if cfg.UseTor {
		return NewTor(cfg.ThrottleCodes)
	}

	return NewStandard(cfg.Threshold)
}

func NewTor(throttleCodes []int) Throttler {
	return NewTorThrottler(throttleCodes)
}

func NewStandard(threshold int) Throttler {
	t := &StandardThrottler{
		threshold:                 threshold,
		backoff:                   5 * time.Minute,
		recoveredSinceLastTrigger: true,
		cond:                      sync.NewCond(&sync.Mutex{}),
		startTime:                 time.Now(),
	}

	t.cond = sync.NewCond(&t.mu)
	return t
}
