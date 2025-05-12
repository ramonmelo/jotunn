package attack

import (
	"io"
	"net/http"
	"strings"

	"github.com/LinharesAron/jotunn/internal/config"
	"github.com/LinharesAron/jotunn/internal/types"
	"github.com/LinharesAron/jotunn/internal/utils"
)

func ExecuteAttempt(client *http.Client, cfg *config.AttackConfig, attempt *types.Attempt) (bool, int, error) {

	values := map[string]string{
		"^USER^": attempt.Username,
		"^PASS^": attempt.Password,
	}

	if cfg.CSRFField != "" {
		statusCode, csrfToken, err := utils.RetrieveCSRFToken(client, cfg.CSRFField, cfg.CSRFSourceURL)
		if err != nil {
			return false, statusCode, err
		}
		values["^CSRF^"] = csrfToken
	}

	payload := utils.ReplacePlaceholders(cfg.Payload, values)
	var req *http.Request
	var err error

	if strings.ToUpper(cfg.Method) == "GET" {
		urlWithQuery := cfg.URL + "?" + payload
		req, err = http.NewRequest(cfg.Method, urlWithQuery, nil)
	} else {
		req, err = http.NewRequest(cfg.Method, cfg.URL, strings.NewReader(payload))
	}

	if err != nil {
		return false, -1, err
	}

	for k, v := range cfg.Headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, -1, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, -1, err
	}

	body := string(bodyBytes)
	statusCode := resp.StatusCode

	if isValidResponse(cfg, body) {
		if isValidStatusCode(statusCode) {
			return true, statusCode, nil
		}
		return false, statusCode, &InvalidStatusCode{statusCode}
	}
	return isValidResponse(cfg, body), statusCode, nil
}

func isValidStatusCode(statusCode int) bool {
	return statusCode >= 200 && statusCode < 400
}

func isValidResponse(cfg *config.AttackConfig, body string) bool {
	success, keyword := cfg.Keyword()
	containsKeyword := strings.Contains(body, keyword)

	return success == containsKeyword
}
