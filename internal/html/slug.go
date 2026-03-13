package html

import (
	"regexp"
	"strings"
)

func Slug(name string) string {
	s := strings.ToLower(name)
	s = regexp.MustCompile(`[\s\pP]+`).ReplaceAllString(s, "-")
	s = regexp.MustCompile(`-+`).ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}
