package spec

import (
	"fmt"
	"strings"

	"github.com/bssm-oss/web-replica/internal/fsutil"
)

func BuildBrief(designSpec DesignSpec) string {
	var b strings.Builder
	b.WriteString("# Siteforge Brief\n\n")
	b.WriteString("## Source\n")
	b.WriteString(fmt.Sprintf("- URL: %s\n", designSpec.SourceURL))
	b.WriteString(fmt.Sprintf("- Mode: %s\n\n", designSpec.Mode))
	b.WriteString("## Page Purpose\n")
	b.WriteString(designSpec.Page.ContentSummary)
	b.WriteString("\n\n## Layout\n")
	b.WriteString(fmt.Sprintf("- Header: %s\n", summarizeSections(designSpec.Page.Structure.Sections, "header")))
	b.WriteString(fmt.Sprintf("- Hero: %s\n", summarizeSections(designSpec.Page.Structure.Sections, "hero")))
	b.WriteString(fmt.Sprintf("- Main sections: %s\n", summarizeMainSections(designSpec.Page.Structure.Sections)))
	b.WriteString(fmt.Sprintf("- Footer: %s\n\n", summarizeSections(designSpec.Page.Structure.Sections, "footer")))
	b.WriteString("## Visual Direction\n")
	b.WriteString(fmt.Sprintf("- Colors: %s\n", strings.Join(append(designSpec.DesignTokens.Colors.Background, designSpec.DesignTokens.Colors.Accent...), ", ")))
	b.WriteString(fmt.Sprintf("- Typography: %s\n", strings.Join(designSpec.DesignTokens.Typography.FontFamilies, ", ")))
	b.WriteString(fmt.Sprintf("- Spacing: %s\n", strings.Join(designSpec.DesignTokens.Spacing, ", ")))
	b.WriteString(fmt.Sprintf("- Cards: %s\n", summarizeSections(designSpec.Page.Structure.Sections, "card")))
	b.WriteString(fmt.Sprintf("- Buttons: %s\n\n", strings.Join(firstNavTexts(designSpec.Page.Structure.Navigation, 4), ", ")))
	b.WriteString("## Responsive Behavior\n")
	b.WriteString(fmt.Sprintf("- Desktop: %s\n", strings.Join(designSpec.Responsive.Desktop.Notes, "; ")))
	b.WriteString(fmt.Sprintf("- Tablet: %s\n", strings.Join(designSpec.Responsive.Tablet.Notes, "; ")))
	b.WriteString(fmt.Sprintf("- Mobile: %s\n\n", strings.Join(designSpec.Responsive.Mobile.Notes, "; ")))
	b.WriteString("## Asset Policy\n")
	b.WriteString("기본적으로 placeholder 사용.\n")
	b.WriteString("사용자 소유 asset만 명시 플래그가 있을 때 허용.\n\n")
	b.WriteString("## Generation Notes\n")
	for _, rule := range designSpec.GenerationRules {
		b.WriteString("- " + rule + "\n")
	}
	return b.String()
}

func WriteBrief(path string, designSpec DesignSpec) error {
	return fsutil.SafeWriteFile(path, []byte(BuildBrief(designSpec)), 0o644)
}

func summarizeSections(sections []Section, kind string) string {
	for _, section := range sections {
		if section.Kind == kind {
			if section.Heading != "" {
				return section.Heading
			}
			if section.Summary != "" {
				return section.Summary
			}
		}
	}
	return "not detected"
}

func summarizeMainSections(sections []Section) string {
	parts := make([]string, 0, len(sections))
	for _, section := range sections {
		if section.Kind == "section" || section.Kind == "hero" || section.Kind == "card-list" {
			label := section.Heading
			if label == "" {
				label = section.Summary
			}
			if label != "" {
				parts = append(parts, label)
			}
		}
		if len(parts) == 4 {
			break
		}
	}
	if len(parts) == 0 {
		return "not detected"
	}
	return strings.Join(parts, ", ")
}

func firstNavTexts(items []NavItem, limit int) []string {
	texts := make([]string, 0, limit)
	for _, item := range items {
		if item.Text != "" {
			texts = append(texts, item.Text)
		}
		if len(texts) == limit {
			break
		}
	}
	if len(texts) == 0 {
		return []string{"placeholder CTA"}
	}
	return texts
}
