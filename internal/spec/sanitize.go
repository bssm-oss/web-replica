package spec

import (
	"regexp"
	"strings"
)

var whitespacePattern = regexp.MustCompile(`\s+`)

func CleanText(value string) string {
	value = whitespacePattern.ReplaceAllString(strings.TrimSpace(value), " ")
	return strings.Trim(value, " \n\t")
}

func ShortText(value string, limit int) string {
	clean := CleanText(value)
	if len(clean) <= limit {
		return clean
	}
	if limit <= 3 {
		return clean[:limit]
	}
	return clean[:limit-3] + "..."
}

func SummarizeText(chunks []string) string {
	parts := make([]string, 0, len(chunks))
	for _, chunk := range chunks {
		clean := ShortText(chunk, 140)
		if clean != "" {
			parts = append(parts, clean)
		}
		if len(parts) == 3 {
			break
		}
	}
	return strings.Join(parts, " ")
}
