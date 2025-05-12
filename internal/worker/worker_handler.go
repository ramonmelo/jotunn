package worker

import (
	"sync"

	"github.com/LinharesAron/jotunn/internal/attack"
	"github.com/LinharesAron/jotunn/internal/config"
	"github.com/LinharesAron/jotunn/internal/httpclient"
	"github.com/LinharesAron/jotunn/internal/logger"
	"github.com/LinharesAron/jotunn/internal/throttle"
	"github.com/LinharesAron/jotunn/internal/types"
	"github.com/LinharesAron/jotunn/internal/utils"
)

type WokerHandler struct {
	throttle throttle.Throttler
	cfg      *config.AttackConfig
}

func NewAttack(cfg *config.AttackConfig, thorttle throttle.Throttler) Worker {
	return &WokerHandler{
		cfg:      cfg,
		throttle: thorttle,
	}
}

func (w *WokerHandler) Start(wg *sync.WaitGroup, input <-chan types.Attempt, shouldRetry func(types.Attempt) error) {
	defer wg.Done()

	client := httpclient.Get()
	for attempt := range input {
		w.throttle.WaitIfBlocked()
		w.throttle.WaitCadence()
		w.throttle.RegisterRequest()

		success, statusCode, err := attack.ExecuteAttempt(client, w.cfg, &attempt)
		if err != nil {
			logger.Progress.AddError()
			if utils.IsTimeoutOrConnectionError(err) || w.cfg.IsThrottlingStatus(statusCode) {
				if err := shouldRetry(attempt); err == nil {
					logger.Progress.AddRetry()
					continue
				}
				logger.Warn("[StandardThrottler] Retry limit reached for %s:%s – ignoring attempt → %s", attempt.Username, attempt.Password, err)
			} else {
				logger.Error("[Worker] Request error: %v\n", err)
			}
		}

		if success {
			logger.Success("🎯 [Worker] [%d] Valid username:password → %s:%s 🎯", statusCode, attempt.Username, attempt.Password)
			logger.Progress.AddSuccess()
		}

		w.throttle.MarkRecovered()
		logger.Progress.Inc()
	}
}
