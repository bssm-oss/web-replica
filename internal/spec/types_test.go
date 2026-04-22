package spec

import (
	"encoding/json"
	"testing"
)

func TestDesignSpecMarshalUnmarshal(t *testing.T) {
	input := DesignSpec{
		SchemaVersion: "0.1",
		SourceURL:     "https://example.com",
		NormalizedURL: "https://example.com/",
		Mode:          "inspired_reimplementation",
		CreatedAt:     "2026-04-22T12:00:00Z",
		Page:          Page{Title: "Example", Description: "Short description", Language: "en", ContentSummary: "Summary"},
	}
	payload, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var output DesignSpec
	if err := json.Unmarshal(payload, &output); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if output.SchemaVersion == "" || output.SourceURL == "" || output.Mode == "" {
		t.Fatalf("required fields missing after roundtrip: %#v", output)
	}
}
