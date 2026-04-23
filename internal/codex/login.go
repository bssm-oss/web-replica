package codex

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

type LoginOptions struct {
	Status     bool
	DeviceAuth bool
}

func BuildLoginArgs(opts LoginOptions) []string {
	args := []string{"login"}
	if opts.Status {
		return append(args, "status")
	}
	if opts.DeviceAuth {
		args = append(args, "--device-auth")
	}
	return args
}

func RunLogin(ctx context.Context, opts LoginOptions) error {
	exe, err := exec.LookPath("codex")
	if err != nil {
		return fmt.Errorf("codex cli not found. Install with:\n\nnpm i -g @openai/codex\ncodex\n")
	}
	cmd := exec.CommandContext(ctx, exe, BuildLoginArgs(opts)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
