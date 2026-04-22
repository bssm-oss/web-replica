package browser

import (
	"context"
	"testing"
)

func TestResolveViewports(t *testing.T) {
	viewports, err := ResolveViewports("desktop,mobile")
	if err != nil {
		t.Fatalf("ResolveViewports error: %v", err)
	}
	if len(viewports) != 2 || viewports[0].Name != "desktop" || viewports[1].Name != "mobile" {
		t.Fatalf("unexpected viewports: %#v", viewports)
	}
	if _, err := ResolveViewports("unknown"); err == nil {
		t.Fatal("expected error for unknown viewport")
	}
}

func TestRelativeScreenshotPath(t *testing.T) {
	got := RelativeScreenshotPath("/tmp/run", "/tmp/run/screenshots/desktop.png")
	if got != "screenshots/desktop.png" {
		t.Fatalf("unexpected relative path %q", got)
	}
}

func TestIsBrowserRequestAllowed(t *testing.T) {
	if !isBrowserRequestAllowed(context.Background(), "data:text/plain,hello", false) {
		t.Fatal("expected data url to be allowed")
	}
	if isBrowserRequestAllowed(context.Background(), "http://127.0.0.1:8080/private", false) {
		t.Fatal("expected loopback request to be blocked")
	}
}
