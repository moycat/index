package snippet

import (
	"strings"
)

type Builder struct{}

func NewBuilder() *Builder {
	return &Builder{}
}

func (b *Builder) Build(content string, terms []string, maxRunes int) string {
	if maxRunes < 1 {
		maxRunes = 120
	}
	runes := []rune(strings.TrimSpace(content))
	if len(runes) == 0 {
		return ""
	}
	if len(runes) <= maxRunes {
		return string(runes)
	}

	start := findStartIndex(runes, terms)
	if start < 0 {
		start = 0
	}

	half := maxRunes / 2
	from := start - half
	if from < 0 {
		from = 0
	}
	to := from + maxRunes
	if to > len(runes) {
		to = len(runes)
		from = to - maxRunes
		if from < 0 {
			from = 0
		}
	}

	prefix := ""
	suffix := ""
	if from > 0 {
		prefix = "..."
	}
	if to < len(runes) {
		suffix = "..."
	}
	return prefix + string(runes[from:to]) + suffix
}

func findStartIndex(runes []rune, terms []string) int {
	if len(terms) == 0 {
		return -1
	}
	lowerContent := strings.ToLower(string(runes))
	contentRunes := []rune(lowerContent)
	best := -1

	for _, term := range terms {
		termRunes := []rune(strings.ToLower(strings.TrimSpace(term)))
		if len(termRunes) == 0 || len(termRunes) > len(contentRunes) {
			continue
		}
		idx := runeIndex(contentRunes, termRunes)
		if idx >= 0 && (best < 0 || idx < best) {
			best = idx
		}
	}

	return best
}

func runeIndex(haystack, needle []rune) int {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if runesEqual(haystack[i:i+len(needle)], needle) {
			return i
		}
	}
	return -1
}

func runesEqual(a, b []rune) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
