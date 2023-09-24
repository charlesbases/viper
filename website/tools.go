package website

import (
	"os"
	"strings"
)

// ReadLineWithFile .
func ReadLineWithFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var links = make([]string, 0)

	return links, ReadLine(func(line string) (isBreak bool) {
		if !strings.HasPrefix(strings.TrimSpace(line), "#") {
			links = append(links, line)
		}
		return false
	})(file)
}
