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
	if !contains(args, `approval_policy="on-request"`) {
		t.Fatalf("expected approval policy config override, got %#v", args)
	}
	if args[len(args)-1] != "make app" {
		t.Fatalf("expected prompt to remain a single final arg, got %#v", args)
	}
}

func TestBuildArgsNeverDoesNotUseFullAuto(t *testing.T) {
	args, err := BuildArgs(CommandOptions{ApprovalMode: "never", OutputDir: "./generated-site", Prompt: "make app"})
	if err != nil {
		t.Fatalf("BuildArgs error: %v", err)
	}
	if contains(args, "--full-auto") {
		t.Fatalf("never mode must not be mapped to --full-auto: %#v", args)
	}
	if !contains(args, `approval_policy="never"`) {
		t.Fatalf("expected never approval policy config override, got %#v", args)
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
