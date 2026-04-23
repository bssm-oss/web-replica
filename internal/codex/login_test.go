package codex

import "testing"

func TestBuildLoginArgs(t *testing.T) {
	tests := []struct {
		name string
		opts LoginOptions
		want []string
	}{
		{name: "default", opts: LoginOptions{}, want: []string{"login"}},
		{name: "status", opts: LoginOptions{Status: true}, want: []string{"login", "status"}},
		{name: "device auth", opts: LoginOptions{DeviceAuth: true}, want: []string{"login", "--device-auth"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildLoginArgs(tt.opts)
			if len(got) != len(tt.want) {
				t.Fatalf("expected %#v, got %#v", tt.want, got)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("expected %#v, got %#v", tt.want, got)
				}
			}
		})
	}
}
