package analyzer

import (
	"net/url"
	"testing"
)

func TestFilterAssetCandidate(t *testing.T) {
	base, _ := url.Parse("https://example.com/")
	tests := []struct {
		name          string
		rawURL        string
		mimeType      string
		allowOwned    bool
		wantAllowed   bool
		wantLocalPath string
	}{
		{name: "same origin allowed", rawURL: "/images/logo.png", mimeType: "image", allowOwned: true, wantAllowed: true, wantLocalPath: "logo.png"},
		{name: "third party blocked", rawURL: "https://cdn.example.net/logo.png", mimeType: "image", allowOwned: true, wantAllowed: false, wantLocalPath: "logo.png"},
		{name: "tracking blocked", rawURL: "https://example.com/pixel-analytics.gif", mimeType: "image", allowOwned: true, wantAllowed: false, wantLocalPath: "pixel-analytics.gif"},
		{name: "sanitized filename", rawURL: "/../../secret.png", mimeType: "image", allowOwned: true, wantAllowed: true, wantLocalPath: "secret.png"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := FilterAssetCandidate(base, tt.rawURL, tt.mimeType, tt.allowOwned)
			if entry.Allowed != tt.wantAllowed {
				t.Fatalf("expected allowed=%v, got %v", tt.wantAllowed, entry.Allowed)
			}
			if entry.LocalPath != tt.wantLocalPath {
				t.Fatalf("expected local path %q, got %q", tt.wantLocalPath, entry.LocalPath)
			}
		})
	}
}
