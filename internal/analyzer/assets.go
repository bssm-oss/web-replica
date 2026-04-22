package analyzer

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/bssm-oss/web-replica/internal/fsutil"
	"github.com/bssm-oss/web-replica/internal/logging"
	"github.com/bssm-oss/web-replica/internal/spec"
)

var blockedAssetKeywords = []string{"google-analytics", "googletagmanager", "facebook", "doubleclick", "hotjar", "segment", "mixpanel", "amplitude", "pixel", "tracker", "ads", "analytics"}

const maxOwnedAssetBytes int64 = 5 << 20

func CollectAssetCandidates(doc *goquery.Document, sourceURL string, allowOwnedAssets bool) []spec.AssetEntry {
	base, _ := url.Parse(sourceURL)
	items := make([]spec.AssetEntry, 0, 16)
	appendAsset := func(rawURL string, mimeType string) {
		entry := FilterAssetCandidate(base, rawURL, mimeType, allowOwnedAssets)
		if entry.URL != "" {
			items = append(items, entry)
		}
	}
	doc.Find("img[src]").Each(func(_ int, s *goquery.Selection) { appendAsset(s.AttrOr("src", ""), "image") })
	doc.Find(`link[rel="preload"][as="font"], link[href*="font"]`).Each(func(_ int, s *goquery.Selection) { appendAsset(s.AttrOr("href", ""), "font") })
	return items
}

func DownloadOwnedAssets(ctx context.Context, assets []spec.AssetEntry, runDir string, logger *logging.Logger) []spec.AssetEntry {
	updated := make([]spec.AssetEntry, 0, len(assets))
	client := &http.Client{}
	for _, asset := range assets {
		if !asset.Allowed {
			updated = append(updated, asset)
			continue
		}
		relDir := filepath.ToSlash(filepath.Join("owned-assets", assetTypeDir(asset.MimeType)))
		relPath := filepath.ToSlash(filepath.Join(relDir, sanitizedAssetName(asset.LocalPath)))
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, asset.URL, nil)
		if err != nil {
			asset.Allowed = false
			asset.Reason = fmt.Sprintf("asset request creation failed: %v", err)
			updated = append(updated, asset)
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			asset.Allowed = false
			asset.Reason = fmt.Sprintf("asset download failed: %v", err)
			updated = append(updated, asset)
			continue
		}
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxOwnedAssetBytes+1))
		_ = resp.Body.Close()
		if readErr != nil {
			asset.Allowed = false
			asset.Reason = fmt.Sprintf("asset read failed: %v", readErr)
			updated = append(updated, asset)
			continue
		}
		if int64(len(body)) > maxOwnedAssetBytes {
			asset.Allowed = false
			asset.Reason = "asset exceeded size limit"
			updated = append(updated, asset)
			continue
		}
		if !mimeTypeMatches(resp.Header.Get("Content-Type"), asset.MimeType) {
			asset.Allowed = false
			asset.Reason = "asset content type did not match policy"
			updated = append(updated, asset)
			continue
		}
		resolved, err := fsutil.SafeJoin(runDir, filepath.FromSlash(relPath))
		if err != nil {
			asset.Allowed = false
			asset.Reason = fmt.Sprintf("asset path rejected: %v", err)
			updated = append(updated, asset)
			continue
		}
		if err := fsutil.SafeWriteFile(resolved, body, 0o644); err != nil {
			asset.Allowed = false
			asset.Reason = fmt.Sprintf("asset write failed: %v", err)
			updated = append(updated, asset)
			continue
		}
		asset.LocalPath = relPath
		asset.Reason = "same-origin owned asset downloaded"
		if logger != nil {
			logger.Verbosef("Downloaded allowed owned asset to %s", relPath)
		}
		updated = append(updated, asset)
	}
	return updated
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

func mimeTypeMatches(header string, want string) bool {
	mediaType, _, err := mime.ParseMediaType(header)
	if err != nil {
		mediaType = header
	}
	mediaType = strings.ToLower(strings.TrimSpace(mediaType))
	switch want {
	case "image":
		return strings.HasPrefix(mediaType, "image/")
	case "font":
		return strings.HasPrefix(mediaType, "font/") || mediaType == "application/font-woff" || mediaType == "application/font-woff2" || mediaType == "application/octet-stream"
	default:
		return false
	}
}

func assetTypeDir(mimeType string) string {
	if mimeType == "font" {
		return "fonts"
	}
	return "images"
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
