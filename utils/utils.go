package utils

import (
	"regexp"
	"strings"
)

func ToID(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "")
	s = strings.TrimSuffix(s, "\n")

	rg := regexp.MustCompile("[^a-z0-9]+")
	s = rg.ReplaceAllString(s, "")
	return s
}
