package website

import (
	"os"
	"regexp"
	"strconv"
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

// FindSubnumber .
func FindSubnumber(r *regexp.Regexp, s string) int64 {
	var subint int64
	if substr := FindSubstring(r, s); len(substr) != 0 {
		subint, _ = strconv.ParseInt(substr, 10, 64)
	}
	return subint
}

// FindSubstring .
func FindSubstring(r *regexp.Regexp, s string) string {
	if subs := r.FindStringSubmatch(s); len(subs) > 1 {
		return subs[1]
	}
	return ""
}
