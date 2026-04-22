package analyzer

import (
	"net/url"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/bssm-oss/web-replica/internal/spec"
)

var blockedAssetKeywords = []string{"google-analytics", "googletagmanager", "facebook", "doubleclick", "hotjar", "segment", "mixpanel", "amplitude", "pixel", "tracker", "ads", "analytics"}

func CollectAssetCandidates(doc *goquery.Document, sourceURL string) []spec.AssetEntry {
	base, _ := url.Parse(sourceURL)
	items := make([]spec.AssetEntry, 0, 16)
	appendAsset := func(rawURL string, mimeType string) {
		entry := FilterAssetCandidate(base, rawURL, mimeType, false)
		if entry.URL != "" {
			items = append(items, entry)
		}
	}
	doc.Find("img[src]").Each(func(_ int, s *goquery.Selection) { appendAsset(s.AttrOr("src", ""), "image") })
	doc.Find(`link[rel="preload"][as="font"], link[href*="font"]`).Each(func(_ int, s *goquery.Selection) { appendAsset(s.AttrOr("href", ""), "font") })
	return items
}

func FilterAssetCandidate(base *url.URL, rawURL string, mimeType string, allowOwnedAssets bool) spec.AssetEntry {
	if strings.TrimSpace(rawURL) == "" {
		return spec.AssetEntry{}
	}
	resolved, err := base.Parse(rawURL)
	if err != nil {
		return spec.AssetEntry{URL: rawURL, MimeType: mimeType, Allowed: false, Reason: "invalid asset url"}
	}
	lower := strings.ToLower(resolved.String())
	for _, keyword := range blockedAssetKeywords {
		if strings.Contains(lower, keyword) {
			return spec.AssetEntry{URL: resolved.String(), MimeType: mimeType, Allowed: false, Reason: "tracking or advertising asset blocked", LocalPath: sanitizedAssetName(resolved.Path)}
		}
	}
	sameOrigin := base != nil && strings.EqualFold(base.Hostname(), resolved.Hostname())
	allowed := allowOwnedAssets && sameOrigin && (mimeType == "image" || mimeType == "font")
	reason := "placeholders used by default"
	if allowed {
		reason = "same-origin owned asset allowed"
	}
	if mimeType == "script" {
		allowed = false
		reason = "script downloads are never allowed"
	}
	return spec.AssetEntry{URL: resolved.String(), MimeType: mimeType, Allowed: allowed, Reason: reason, LocalPath: sanitizedAssetName(resolved.Path)}
}

func sanitizedAssetName(assetPath string) string {
	base := filepath.Base(filepath.Clean(assetPath))
	base = strings.ReplaceAll(base, "..", "")
	base = strings.ReplaceAll(base, string(filepath.Separator), "-")
	if base == "." || base == "" {
		return "asset"
	}
	return base
}
