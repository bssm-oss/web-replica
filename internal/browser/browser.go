package browser

import "fmt"

type Viewport struct {
	Name   string
	Width  int64
	Height int64
}

var presets = map[string]Viewport{
	"desktop": {Name: "desktop", Width: 1440, Height: 1000},
	"tablet":  {Name: "tablet", Width: 768, Height: 1000},
	"mobile":  {Name: "mobile", Width: 390, Height: 844},
}

func ResolveViewports(selection string) ([]Viewport, error) {
	if selection == "all" {
		return []Viewport{presets["desktop"], presets["tablet"], presets["mobile"]}, nil
	}
	if selection == "desktop,tablet,mobile" || selection == "" {
		return []Viewport{presets["desktop"], presets["tablet"], presets["mobile"]}, nil
	}
	items := []Viewport{}
	seen := map[string]bool{}
	for _, part := range splitSelection(selection) {
		viewport, ok := presets[part]
		if !ok {
			return nil, fmt.Errorf("unsupported viewport %q", part)
		}
		if !seen[part] {
			items = append(items, viewport)
			seen[part] = true
		}
	}
	return items, nil
}

func splitSelection(selection string) []string {
	parts := []string{}
	current := ""
	for _, r := range selection {
		if r == ',' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
			continue
		}
		if r != ' ' {
			current += string(r)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}
