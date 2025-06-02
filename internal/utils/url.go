package utils

import (
	"net/url"
	"strings"
)

func GetDomain(URL string) (string, error) {
	parsedURL, err := url.Parse(URL)
	if err != nil {
		return "", err
	}

	host := parsedURL.Host
	if strings.Contains(host, ":") {
		parts := strings.Split(host, ":")
		host = parts[0]
	}
	return host, nil
}
