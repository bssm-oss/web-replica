package analyzer

import (
	"context"
	"testing"
)

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "https url ok", input: "https://1.1.1.1", wantErr: false},
		{name: "http url ok", input: "http://8.8.8.8", wantErr: false},
		{name: "empty reject", input: "", wantErr: true},
		{name: "javascript reject", input: "javascript:alert(1)", wantErr: true},
		{name: "file reject", input: "file:///tmp/demo", wantErr: true},
		{name: "localhost reject", input: "http://localhost:3000", wantErr: true},
		{name: "private ip reject", input: "http://192.168.0.1", wantErr: true},
		{name: "malformed reject", input: "://bad-url", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateURL(context.Background(), tt.input)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error for %q", tt.input)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.input, err)
			}
		})
	}
}
