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
		return `High-fidelity safe clone-coding mode:
- Treat screenshots as the primary visual reference.
- Match the source page's composition, section order, viewport proportions, whitespace rhythm, typography scale, color palette, border/radius language, and responsive breakpoints as closely as possible.
- Recreate phone/device mockups, cards, hero hierarchy, CTA placement, footer density, and scroll rhythm with original local CSS/React code.
- Do not copy protected logos, brand names, long original text, copyrighted images, source code, tracking scripts, or third-party assets.
- Replace protected brand marks and copy with neutral equivalents while preserving the same visual role and layout weight.
- Prefer CSS gradients, local SVG placeholders, and recreated UI shapes over downloaded source assets.`
	}
	return `Standard safe reimplementation mode:
- Use the source as directional design input.
- Preserve the broad layout and responsive intent without trying to match every pixel.
- Do not copy protected logos, brand names, long text, images, source code, or tracking scripts.`
}
