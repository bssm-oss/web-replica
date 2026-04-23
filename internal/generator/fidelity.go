package generator

import (
	"fmt"
	"strings"
)

const (
	FidelityStandard = "standard"
	FidelityHigh     = "high"
)

func NormalizeFidelity(value string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return FidelityStandard, nil
	}
	switch normalized {
	case FidelityStandard, FidelityHigh:
		return normalized, nil
	default:
		return "", fmt.Errorf("unsupported fidelity mode %q; allowed values: standard, high", value)
	}
}

func FidelityGuidance(mode string) string {
	if mode == FidelityHigh {
		return `High-fidelity clone-coding mode. Goal: make the result visually indistinguishable from the original at a glance.

PRIORITY ORDER (most important first):
1. Section order & count — every section from the original must appear in the same order.
2. Spacing & whitespace — section top/bottom padding must be generous (120px–250px). Do not compress sections.
3. Color palette — use exact hex values from DesignSpec. White (#fff) and light gray (#f9fafb) alternate between sections. Brand blue used only for CTAs, labels, and accents.
4. Typography — headings: weight 800–900, tight letter-spacing (−0.04em to −0.06em), large fluid sizes (clamp). Body: weight 600–700. Use the font families listed in DesignSpec verbatim.
5. Component fidelity — phone mockups must have dark border, rounded corners (44px), perspective tilt, and realistic app UI content inside. Hero must have a large centered phone with floating feature-label tiles around it.
6. CTA buttons — pill-shaped (border-radius: 9999px), brand-blue fill, white text, box-shadow glow.
7. Photography/illustration sections — where the original has real photos (person with phone, lifestyle), render a full-width gradient placeholder block with matching dominant color tone (e.g. light blue, warm cream).
8. Dark sections — business / B2B sections use very dark background (#0f172a or #111827) with white text and subtle glass-card borders.
9. Footer — 4-column dark footer (#191f28) with company links, small legal text, and brand wordmark.
10. Responsive — desktop (≥1200px) side-by-side grids; tablet (768–1199px) 2-col or stacked; mobile (<768px) fully stacked, larger touch targets.

CRAFT DETAILS:
- backdrop-filter: blur on fixed header.
- Floating tiles: absolute-positioned, subtle box-shadow, slight rotateX perspective.
- Section titles have a small blue eyebrow label above the main heading.
- Transfer/feature sections alternate: copy-left + visual-right, then visual-left + copy-right.
- All Korean text must exactly match the source content from RawHTML and BriefMarkdown.
- No placeholder "Lorem ipsum" — use the actual Korean copy from the source.`
	}
	return `Standard reimplementation mode:
- Use the source as directional design input.
- Preserve the broad layout and responsive intent without trying to match every pixel.`
}
