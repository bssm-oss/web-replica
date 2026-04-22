package logging

import (
	"strings"
	"testing"
)

func TestRedactSecrets(t *testing.T) {
	input := "Authorization: Bearer abc\nCookie: hello=world\nOPENAI_API_KEY=sk-test\nheader eyJhbGciOiJIUzI1NiJ9.payload.signature1234567890"
	masked := RedactSecrets(input)
	if masked == input {
		t.Fatal("expected secrets to be redacted")
	}
	if strings.Contains(masked, "sk-test") || strings.Contains(masked, "hello=world") || strings.Contains(masked, "signature1234567890") {
		t.Fatalf("sensitive values still present: %q", masked)
	}
}
