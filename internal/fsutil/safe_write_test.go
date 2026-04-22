package fsutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSafeWriteRelative(t *testing.T) {
	base := t.TempDir()
	path, err := SafeWriteRelative(base, "nested/file.txt", []byte("hello"), 0o644)
	if err != nil {
		t.Fatalf("SafeWriteRelative error: %v", err)
	}
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if string(payload) != "hello" {
		t.Fatalf("unexpected payload %q", string(payload))
	}
	_, err = SafeWriteRelative(base, filepath.Join("..", "escape.txt"), []byte("bad"), 0o644)
	if err == nil {
		t.Fatal("expected path traversal rejection")
	}
}
