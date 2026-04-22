package analyzer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/bssm-oss/web-replica/internal/spec"
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

func TestDownloadOwnedAssets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/images/logo.png" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write([]byte("png-data"))
	}))
	defer server.Close()
	assets := []spec.AssetEntry{{URL: server.URL + "/images/logo.png", MimeType: "image", Allowed: true, LocalPath: "logo.png", Reason: "same-origin owned asset allowed"}}
	runDir := t.TempDir()
	updated := DownloadOwnedAssets(context.Background(), assets, runDir, nil)
	if len(updated) != 1 {
		t.Fatalf("expected one asset, got %d", len(updated))
	}
	if !updated[0].Allowed {
		t.Fatalf("expected asset to remain allowed: %#v", updated[0])
	}
	if updated[0].LocalPath == "" {
		t.Fatalf("expected downloaded local path, got %#v", updated[0])
	}
	if _, err := os.Stat(filepath.Join(runDir, filepath.FromSlash(updated[0].LocalPath))); err != nil {
		t.Fatalf("expected downloaded asset file to exist: %v", err)
	}
}
