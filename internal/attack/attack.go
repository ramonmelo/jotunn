package attack

import (
	"io"
	"net/http"
	"strings"

	"github.com/LinharesAron/jotunn/internal/config"
	"github.com/LinharesAron/jotunn/internal/types"
	"github.com/LinharesAron/jotunn/internal/utils"
)

func ExecuteAttempt(client *http.Client, cfg *config.AttackConfig, attempt types.Attempt) (bool, int, error) {

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

	payload, err := utils.SafeReplacePayload(cfg.Payload, values)
	if err != nil {
		return false, -1, err
	}

	var req *http.Request

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

	valid, err := isValidResponse(cfg, body, statusCode)
	if err != nil {
		return false, statusCode, err
	}
	return valid, statusCode, nil
}

func isValidStatusCode(statusCode int) bool {
	return statusCode >= 200 && statusCode < 400
}

func isValidResponse(cfg *config.AttackConfig, body string, statusCode int) (bool, error) {
	keyword := cfg.Keyword()
	containsKeyword := strings.Contains(body, keyword)

	if cfg.IsSuccessKeyword {
		return containsKeyword, nil
	}

	if containsKeyword {
		return false, nil
	}

	validStatusCode := isValidStatusCode(statusCode)
	if !validStatusCode {
		return false, &InvalidStatusCode{statusCode}
	}

	return true, nil
}
