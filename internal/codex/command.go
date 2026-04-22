package codex

import (
	"errors"
	"fmt"
	"strings"
)

type CommandOptions struct {
	ApprovalMode string
	Model        string
	OutputDir    string
	Prompt       string
}

func BuildArgs(opts CommandOptions) ([]string, error) {
	if opts.Prompt == "" {
		return nil, errors.New("prompt is required")
	}
	args := []string{
		"exec",
		"--skip-git-repo-check",
		"--ephemeral",
		"--disable", "plugins",
		"--disable", "apps",
		"--disable", "multi_agent",
		"-c", `model_reasoning_effort="medium"`,
		"-c", `developer_instructions=""`,
		"-c", `mcp_servers={}`,
	}
	switch opts.ApprovalMode {
	case "", "on-request":
		args = append(args, "-c", `approval_policy="on-request"`, "--sandbox", "workspace-write")
	case "untrusted":
		args = append(args, "-c", `approval_policy="untrusted"`, "--sandbox", "workspace-write")
	case "never":
		args = append(args, "-c", `approval_policy="never"`, "--sandbox", "workspace-write")
	default:
		return nil, fmt.Errorf("unsupported codex approval mode %q", opts.ApprovalMode)
	}
	args = append(args, "-C", opts.OutputDir)
	if strings.TrimSpace(opts.Model) != "" {
		args = append(args, "--model", strings.TrimSpace(opts.Model))
	}
	args = append(args, opts.Prompt)
	return args, nil
}
