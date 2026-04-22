package codex

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"

	"github.com/bssm-oss/web-replica/internal/fsutil"
	"github.com/bssm-oss/web-replica/internal/logging"
)

type RunOptions struct {
	OutputDir    string
	RunDir       string
	Prompt       string
	PromptPath   string
	ApprovalMode string
	Model        string
	Logger       *logging.Logger
}

type RunResult struct {
	Stdout  string
	Stderr  string
	LogPath string
}

func Run(ctx context.Context, opts RunOptions) (RunResult, error) {
	exe, err := exec.LookPath("codex")
	if err != nil {
		return RunResult{}, fmt.Errorf("codex cli not found. Install with:\n\nnpm i -g @openai/codex\ncodex\n")
	}
	absOutputDir, err := filepath.Abs(opts.OutputDir)
	if err != nil {
		return RunResult{}, err
	}
	if err := fsutil.EnsureDir(absOutputDir); err != nil {
		return RunResult{}, err
	}
	if opts.PromptPath != "" {
		if err := fsutil.SafeWriteFile(opts.PromptPath, []byte(opts.Prompt), 0o644); err != nil {
			return RunResult{}, err
		}
	}
	args, err := BuildArgs(CommandOptions{ApprovalMode: opts.ApprovalMode, Model: opts.Model, OutputDir: absOutputDir, Prompt: opts.Prompt})
	if err != nil {
		return RunResult{}, err
	}
	cmd := exec.CommandContext(ctx, exe, args...)
	cmd.Dir = absOutputDir
	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(&stdoutBuf)
	cmd.Stderr = io.MultiWriter(&stderrBuf)
	err = cmd.Run()
	logPath := filepath.Join(opts.RunDir, "codex-output.log")
	logPayload := "STDOUT\n" + logging.RedactSecrets(stdoutBuf.String()) + "\n\nSTDERR\n" + logging.RedactSecrets(stderrBuf.String()) + "\n"
	if writeErr := fsutil.SafeWriteFile(logPath, []byte(logPayload), 0o644); writeErr != nil {
		return RunResult{}, writeErr
	}
	if opts.Logger != nil {
		opts.Logger.Verbosef("Ran Codex executable %s with %d args", filepath.Base(exe), len(args))
	}
	if err != nil {
		return RunResult{Stdout: stdoutBuf.String(), Stderr: stderrBuf.String(), LogPath: logPath}, fmt.Errorf("codex exec failed: %w", err)
	}
	return RunResult{Stdout: stdoutBuf.String(), Stderr: stderrBuf.String(), LogPath: logPath}, nil
}
