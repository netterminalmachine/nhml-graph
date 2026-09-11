package helpers

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

const allowed = "abcdefghijklmnopqrstuvwxyz0123456789-_ "

func IsBlank(str string) bool {
	return len(strings.TrimSpace(str)) == 0
}

func SanitizeMigrationName(migName string) string {
	runes := []rune{}

	// decompose into (base letter, accent) pairs. ref: https://go.dev/blog/normalization
	for _, c := range norm.NFD.String(migName) {
		if unicode.Is(unicode.Mn, c) {
			continue // skip nonspacing marks like accents
		}
		nextChar := strings.ToLower(string(c))
		if strings.Contains(allowed, nextChar) {
			runes = append(runes, c)
		}
	}

	return strings.Join(strings.Split(string(runes), " "), "-")
}
