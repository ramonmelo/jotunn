package io

import (
	"bufio"
	"os"
	"strings"
)

// ReadLines reads a file specified by the given path and returns a slice of non-empty,
// trimmed lines as strings. It ignores empty lines. If an error occurs while opening
// or reading the file, it returns the error.
func ReadLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text != "" {
			lines = append(lines, text)
		}
	}
	return lines, scanner.Err()
}
