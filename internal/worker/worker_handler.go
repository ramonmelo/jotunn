package worker

import (
	"sync"

	"github.com/LinharesAron/jotunn/internal/attack"
	"github.com/LinharesAron/jotunn/internal/config"
	"github.com/LinharesAron/jotunn/internal/httpclient"
	"github.com/LinharesAron/jotunn/internal/logger"
	"github.com/LinharesAron/jotunn/internal/throttle"
	"github.com/LinharesAron/jotunn/internal/types"
	"github.com/LinharesAron/jotunn/internal/ui"
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

func (w *WokerHandler) Start(id int, wg *sync.WaitGroup, input <-chan types.Attempt, shouldRetry func(types.Attempt) error) {
	defer wg.Done()

	client := httpclient.Get()
	for attempt := range input {
		w.throttle.Wait()

		success, statusCode, err := attack.ExecuteAttempt(client, w.cfg, attempt)
		if err != nil {
			ui.GetUI().SendProgressEvent(ui.Error)
			if utils.IsTimeoutOrConnectionError(err) || w.cfg.IsThrottlingStatus(statusCode) {
				w.throttle.Trigger()
				if err := shouldRetry(attempt); err == nil {
					ui.GetUI().SendProgressEvent(ui.Retry)
					continue
				}
				logger.Warn("[Worker %d] Retry limit reached for %s:%s – ignoring attempt → %s", id, attempt.Username, attempt.Password, err)
			} else {
				logger.Error("[Worker %d] Request error in the attempt %s:%s -> %v", id, attempt.Username, attempt.Password, err)
			}
		}

		if success {
			logger.Success("🎯 [Worker %d] [%d] Valid username:password → %s:%s 🎯", id, statusCode, attempt.Username, attempt.Password)
			ui.GetUI().SendProgressEvent(ui.Success)
		}

		w.throttle.MarkRecovered()
		ui.GetUI().SendProgressEvent(ui.Inc)
	}
}
