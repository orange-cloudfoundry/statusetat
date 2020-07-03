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
