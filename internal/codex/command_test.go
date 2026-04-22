package codex

import (
	"testing"
)

func TestBuildArgs(t *testing.T) {
	args, err := BuildArgs(CommandOptions{ApprovalMode: "on-request", OutputDir: "./generated-site", Prompt: "make app", Model: "gpt-5"})
	if err != nil {
		t.Fatalf("BuildArgs error: %v", err)
	}
	joined := ""
	for _, arg := range args {
		joined += arg + "|"
	}
	if joined == "" {
		t.Fatal("expected args to be populated")
	}
	if contains(args, "--dangerously-bypass-approvals-and-sandbox") || contains(args, "--yolo") {
		t.Fatal("dangerous mode flags must not be present by default")
	}
	if args[len(args)-1] != "make app" {
		t.Fatalf("expected prompt to remain a single final arg, got %#v", args)
	}
}

func contains(items []string, needle string) bool {
	for _, item := range items {
		if item == needle {
			return true
		}
	}
	return false
}
