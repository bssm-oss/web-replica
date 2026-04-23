package generator

import (
	"strings"
	"testing"
)

func TestNormalizeFidelity(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "default", input: "", want: FidelityStandard},
		{name: "standard", input: "standard", want: FidelityStandard},
		{name: "high", input: " HIGH ", want: FidelityHigh},
		{name: "invalid", input: "exact", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeFidelity(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestFidelityGuidanceHighStaysSafe(t *testing.T) {
	guidance := FidelityGuidance(FidelityHigh)
	for _, phrase := range []string{"Match", "screenshots", "Do not copy protected logos", "neutral equivalents"} {
		if !strings.Contains(guidance, phrase) {
			t.Fatalf("expected high fidelity guidance to contain %q: %s", phrase, guidance)
		}
	}
}
