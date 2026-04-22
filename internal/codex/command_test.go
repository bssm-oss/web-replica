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
	if !contains(args, "--ephemeral") {
		t.Fatalf("expected ephemeral mode for isolated non-interactive runs, got %#v", args)
	}
	if !contains(args, `model_reasoning_effort="medium"`) {
		t.Fatalf("expected bounded reasoning effort for non-interactive runs, got %#v", args)
	}
	if !contains(args, `developer_instructions=""`) {
		t.Fatalf("expected local developer instructions to be cleared, got %#v", args)
	}
	if !contains(args, `mcp_servers={}`) {
		t.Fatalf("expected local mcp server override to be cleared, got %#v", args)
	}
	for _, feature := range []string{"plugins", "apps", "multi_agent"} {
		if !containsPair(args, "--disable", feature) {
			t.Fatalf("expected %q to be disabled for deterministic runs, got %#v", feature, args)
		}
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
		t.Fatalf("expected never approval mode config override, got %#v", args)
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

func containsPair(items []string, left string, right string) bool {
	for i := 0; i < len(items)-1; i++ {
		if items[i] == left && items[i+1] == right {
			return true
		}
	}
	return false
}
