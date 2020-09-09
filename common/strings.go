package common

import (
	"strings"
	"unicode"
)

func Title(content string) string {
	content = strings.TrimSpace(content)
	if len(content) == 0 {
		return ""
	}
	return string(unicode.ToUpper(rune(content[0]))) + content[1:]
}

func InStrSlice(elem string, s []string) bool {
	for _, v := range s {
		if v == elem {
			return true
		}
	}
	return false
}

func FilterStrSlice(elem string, s []string) []string {
	final := make([]string, 0)
	for _, v := range s {
		if v == elem {
			continue
		}
		final = append(final, v)
	}
	return final
}
