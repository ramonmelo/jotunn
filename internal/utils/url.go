package utils

import (
	"net/url"
	"strings"
)

// GetDomain extracts the domain name from the given URL string.
// It parses the URL, removes any port information if present, and returns the domain.
// If the URL cannot be parsed, an error is returned.
//
// Example:
//
//	domain, err := GetDomain("https://example.com:8080/path")
//	// domain == "example.com"
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
