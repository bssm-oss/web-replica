package preview

import "testing"

func TestPreviewCommands(t *testing.T) {
	commands := previewCommands(4173)
	if len(commands) != 2 {
		t.Fatalf("expected two preview commands, got %d", len(commands))
	}
	if commands[0][2] != "preview" || commands[1][2] != "dev" {
		t.Fatalf("unexpected commands: %#v", commands)
	}
}

func TestBytesTrim(t *testing.T) {
	trimmed := string(bytesTrim([]byte("hello\n\r ")))
	if trimmed != "hello" {
		t.Fatalf("unexpected trimmed output %q", trimmed)
	}
}
