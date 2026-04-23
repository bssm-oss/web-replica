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
5. Component fidelity — phone mockups must have dark border, rounded corners (44px), perspective tilt, and realistic app UI content inside. Hero must have a large phone with floating feature-label tiles around it.
6. CTA buttons — pill-shaped (border-radius: 9999px), brand-blue fill, white text, box-shadow glow.
7. Photography/illustration sections — ALWAYS use the actual downloaded images from OwnedAssets when available. Check OwnedAssets for 'section2_X', 'section1_X' files and use them as <img> or CSS background-image. Only fall back to a gradient placeholder if no matching image was downloaded. Do NOT replace real photos with gradients when the image file exists.
8. Dark sections — business / B2B sections use very dark background (#0f172a or #111827) with white text and subtle glass-card borders.
9. Footer — 4-column dark footer (#191f28) with company links, small legal text, and brand wordmark.
10. Responsive — desktop (≥1200px) side-by-side grids; tablet (768–1199px) 2-col or stacked; mobile (<768px) fully stacked, larger touch targets.

CRAFT DETAILS:
- backdrop-filter: blur on fixed header.
- Floating tiles: absolute-positioned, subtle box-shadow, slight rotateX perspective.
- Section titles have a small blue eyebrow label above the main heading.
- Transfer/feature sections alternate: copy-left + visual-right, then visual-left + copy-right.
- All Korean text must exactly match the source content from RawHTML and BriefMarkdown.
- No placeholder "Lorem ipsum" — use the actual Korean copy from the source.

HERO LAYOUT — MANDATORY AT DESKTOP (≥900px):
- The hero section MUST use a 2-column CSS grid: left column = eyebrow + h1 + body copy + CTA pill button; right column = 3D scene stage.
- DO NOT center everything in a single column. Left-align the copy in the left column.
- Hero background: nearly white with very subtle radial gradient. Use ONLY:
    background: radial-gradient(ellipse at 72% 50%, rgba(219,234,254,0.65) 0%, rgba(239,246,255,0.4) 40%, transparent 65%), #ffffff;
  Do NOT use dark blue, saturated blue, or deeply colored backgrounds in the hero.
- Right column: create a relative-positioned stage div (height ~580px). Layer order inside (back to front):
  (a) 3D scene bg: if 'main'/'scene'/'3d'/'object' image downloaded → absolute <img>, width 130%, left -15%, top 50% translateY(-50%), z-index 1, opacity 0.88, pointer-events none.
  (b) Phone composite (MOST IMPORTANT — must be clearly visible, centered and large):
      - If clay/transparent frame downloaded ('clay'/'iPhone'/'frame') → outer <img> as phone frame, width ~230px, height ~460px, centered in stage (left 50% translateX(-50%)), top ~5%, z-index 2, transform rotateX(6deg) rotateZ(-2deg).
      - App screenshot goes INSIDE the frame: position absolute, top 19%, left 20%, width 60%, border-radius 28px, z-index 2 (same stacking context as frame, place before frame in DOM so frame renders on top).
      - The phone must appear large and prominent — do NOT let the 3D scene background dominate.
  (c) Floating label tiles: 4 tiles, position:absolute around the phone, z-index 3, rotateX(8-12deg) rotateZ(±5-9deg), white semi-transparent background (rgba(255,255,255,0.85)), blue box-shadow, backdrop-filter blur.
- If no clay frame image exists, fall back to a CSS-drawn phone div: width 200px, height 400px, border-radius 44px, border 10px solid #1f2937, overflow hidden, with the app screenshot as <img> inside.
- Mobile (<900px): stack to single column, center text, phone stage below copy, stage width ~min(420px,100%).

SECTION IMAGE USAGE — MANDATORY:
- Images with 'section1_2_01', 'section1_2_02', 'section1_2_03' in filename → use in the Home/Consumption feature rows as the visual element (right or left column image, border-radius 28px, width 100%).
- Images with 'section2_1', 'section2_2', 'section2_3' or 'insu', 'document', 'apt' in filename → use as full-width lifestyle photo backgrounds or prominent <img> in the corresponding feature sections, NOT as small thumbnails. These are real lifestyle photos (person holding phone, etc.) — use them at full size with object-fit: cover.
- Image with 'section4_device' or 'device' → use in the business/dark section as a device mockup image on one side.
- Image with 'invest_screen' → use inside a phone mockup in the investment section.
- Image with 'checkout' → use inside a phone mockup in the payment section.
- NEVER replace a downloaded image with a CSS gradient if the file exists in OwnedAssets.`
	}
	return `Standard reimplementation mode:
- Use the source as directional design input.
- Preserve the broad layout and responsive intent without trying to match every pixel.`
}
