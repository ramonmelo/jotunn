package utils

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
)

func RetrieveCSRFToken(client *http.Client, csrfField string, csrfSourceUrl string) (string, error) {
	req, err := http.NewRequest("GET", csrfSourceUrl, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	body := string(bodyBytes)

	re := regexp.MustCompile(`(?i)<input[^>]*name=["']?` + regexp.QuoteMeta(csrfField) + `["']?[^>]*value=["']?([^"'>]+)["']?`)
	match := re.FindStringSubmatch(body)
	if len(match) > 1 {
		return match[1], nil
	}
	return "", fmt.Errorf("CSRF field '%s' not found in HTML body", csrfField)
}
