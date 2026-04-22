package browser

import (
	"sort"
	"strings"

	"github.com/bssm-oss/web-replica/internal/spec"
)

func BuildDesignTokens(captures []ViewportCapture) spec.DesignTokens {
	colorsBg := map[string]int{}
	colorsText := map[string]int{}
	colorsAccent := map[string]int{}
	borders := map[string]int{}
	fontFamilies := map[string]int{}
	fontSizes := map[string]int{}
	fontWeights := map[string]int{}
	lineHeights := map[string]int{}
	spacing := map[string]int{}
	radii := map[string]int{}
	shadows := map[string]int{}
	containerWidths := map[string]int{}
	gridPatterns := map[string]int{}
	flexPatterns := map[string]int{}
	for _, capture := range captures {
		for _, sample := range capture.Samples {
			accumulate(colorsBg, normalizeValue(sample.BackgroundColor))
			accumulate(colorsText, normalizeValue(sample.Color))
			if sample.Selector == "button" || strings.Contains(sample.Selector, "hero") {
				accumulate(colorsAccent, normalizeValue(sample.BackgroundColor))
			}
			accumulate(borders, normalizeValue(sample.Border))
			accumulate(fontFamilies, normalizeFontFamily(sample.FontFamily))
			accumulate(fontSizes, normalizeValue(sample.FontSize))
			accumulate(fontWeights, normalizeValue(sample.FontWeight))
			accumulate(lineHeights, normalizeValue(sample.LineHeight))
			accumulate(spacing, normalizeSpacing(sample.Padding))
			accumulate(spacing, normalizeSpacing(sample.Margin))
			accumulate(radii, normalizeValue(sample.BorderRadius))
			accumulate(shadows, normalizeValue(sample.BoxShadow))
			accumulate(containerWidths, normalizeValue(sample.MaxWidth))
			accumulate(gridPatterns, normalizeValue(sample.GridTemplateColumns))
			accumulate(flexPatterns, normalizeValue(sample.FlexDirection))
		}
	}
	return spec.DesignTokens{
		Colors: spec.ColorTokens{
			Background: topValues(colorsBg, 6),
			Text:       topValues(colorsText, 6),
			Accent:     topValues(colorsAccent, 4),
			Border:     topValues(borders, 4),
		},
		Typography: spec.TypographyTokens{
			FontFamilies: topValues(fontFamilies, 4),
			FontSizes:    topValues(fontSizes, 6),
			FontWeights:  topValues(fontWeights, 6),
			LineHeights:  topValues(lineHeights, 6),
		},
		Spacing: topValues(spacing, 6),
		Radii:   topValues(radii, 4),
		Shadows: topValues(shadows, 4),
		Layout: spec.LayoutTokens{
			ContainerWidths: topValues(containerWidths, 4),
			GridPatterns:    topValues(gridPatterns, 4),
			FlexPatterns:    topValues(flexPatterns, 4),
		},
	}
}

func accumulate(target map[string]int, value string) {
	if value == "" || value == "none" || value == "rgba(0, 0, 0, 0)" || value == "normal" || value == "auto" {
		return
	}
	target[value]++
}

func topValues(input map[string]int, limit int) []string {
	type pair struct {
		Key   string
		Count int
	}
	pairs := make([]pair, 0, len(input))
	for key, count := range input {
		pairs = append(pairs, pair{Key: key, Count: count})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].Count == pairs[j].Count {
			return pairs[i].Key < pairs[j].Key
		}
		return pairs[i].Count > pairs[j].Count
	})
	values := make([]string, 0, min(limit, len(pairs)))
	for _, item := range pairs {
		values = append(values, item.Key)
		if len(values) == limit {
			break
		}
	}
	return values
}

func normalizeValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return strings.Join(strings.Fields(value), " ")
}

func normalizeSpacing(value string) string {
	value = normalizeValue(value)
	if value == "0px" || value == "0px 0px 0px 0px" {
		return ""
	}
	return value
}

func normalizeFontFamily(value string) string {
	value = normalizeValue(value)
	if value == "" {
		return ""
	}
	parts := strings.Split(value, ",")
	if len(parts) == 0 {
		return value
	}
	return strings.TrimSpace(parts[0])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
