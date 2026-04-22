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
	args := []string{"exec", "--skip-git-repo-check"}
	switch opts.ApprovalMode {
	case "", "on-request":
		args = append(args, "--sandbox", "workspace-write")
	case "untrusted":
		args = append(args, "--sandbox", "read-only")
	case "never":
		args = append(args, "--full-auto")
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
