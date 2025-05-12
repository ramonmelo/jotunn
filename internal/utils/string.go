package utils

import "strings"

func ReplacePlaceholders(template string, values map[string]string) string {
	for key, val := range values {
		template = strings.ReplaceAll(template, key, val)
	}
	return template
}

func TruncateAndClean(s string, limit int) string {
	if len(s) > limit {
		s = s[:limit]
	}
	return RemoveNewlines(s)
}

func RemoveNewlines(s string) string {

	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	return s
}
