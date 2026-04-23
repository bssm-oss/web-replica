package analyzer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAnalyzeHTML(t *testing.T) {
	payload, err := os.ReadFile(filepath.Join("..", "..", "testdata", "simple-page.html"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	result, err := AnalyzeHTML(payload, "https://example.com", false)
	if err != nil {
		t.Fatalf("AnalyzeHTML returned error: %v", err)
	}
	if result.Title != "Example Product" {
		t.Fatalf("unexpected title: %q", result.Title)
	}
	if result.ContentSummary == "" || result.ContentSummary == "Create modern landing pages without copying protected content." {
		t.Fatalf("expected summarized content purpose, got %q", result.ContentSummary)
	}
	if len(result.Structure.Headings) == 0 || result.Structure.Headings[0].Text != "Launch better product sites" {
		t.Fatalf("expected heading extraction, got %#v", result.Structure.Headings)
	}
	if len(result.Structure.Navigation) != 2 {
		t.Fatalf("expected navigation links, got %d", len(result.Structure.Navigation))
	}
	if len(result.Structure.Forms) != 1 || len(result.Structure.Forms[0].Fields) == 0 {
		t.Fatalf("expected form fields, got %#v", result.Structure.Forms)
	}
	if len(result.Structure.Images) != 1 || result.Structure.Images[0].Alt == "" {
		t.Fatalf("expected image alt extraction, got %#v", result.Structure.Images)
	}
}

func TestAnalyzeHTMLEmptyCollectionsStayNonNil(t *testing.T) {
	result, err := AnalyzeHTML([]byte(`<html lang="en"><head><title>Only Heading</title></head><body><main><h1>Hello</h1></main></body></html>`), "https://example.com", false)
	if err != nil {
		t.Fatalf("AnalyzeHTML returned error: %v", err)
	}
	if result.Structure.Landmarks == nil || result.Structure.Navigation == nil || result.Structure.Sections == nil || result.Structure.Forms == nil || result.Structure.Images == nil {
		t.Fatalf("expected empty collections to be initialized, got %#v", result.Structure)
	}
}
