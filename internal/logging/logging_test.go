package logging

import (
	"bytes"
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

func TestRedactingWriterFlushesRedactedOutput(t *testing.T) {
	var out bytes.Buffer
	writer := NewRedactingWriter(&out)
	if _, err := writer.Write([]byte("Authorization: Bearer secret-token")); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := writer.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}
	if strings.Contains(out.String(), "secret-token") {
		t.Fatalf("expected output to be redacted, got %q", out.String())
	}
}
