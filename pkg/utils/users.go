package utils

import (
	"regexp"
	"strings"
)

var rankRegexp = regexp.MustCompile(`^[\?\! \+\^\%\@\#\~]`)

func ParseUsername(username string) (name string, rank rune, busy bool) {
	if rankRegexp.MatchString(username) {
		rank = rune(username[0])
		username = username[1:]
	}

	busy = strings.HasSuffix(username, "@!")
	if busy {
		name = username[0 : len(username)-2]
	} else {
		name = username[0:]
	}
	return
}
