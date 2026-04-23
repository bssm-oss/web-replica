package codex

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

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
	Stdout             string
	Stderr             string
	LogPath            string
	ActualApprovalMode string
}

var approvalLinePattern = regexp.MustCompile(`(?m)^approval:\s*([^\n]+)$`)

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
	stdoutWriter := logging.NewRedactingWriter(os.Stdout)
	stderrWriter := logging.NewRedactingWriter(os.Stderr)
	cmd.Stdout = io.MultiWriter(&stdoutBuf, stdoutWriter)
	cmd.Stderr = io.MultiWriter(&stderrBuf, stderrWriter)
	err = cmd.Run()
	_ = stdoutWriter.Flush()
	_ = stderrWriter.Flush()
	actualApproval := detectApprovalMode(stderrBuf.String())
	logPath := filepath.Join(opts.RunDir, "codex-output.log")
	logPayload := "STDOUT\n" + logging.RedactSecrets(stdoutBuf.String()) + "\n\nSTDERR\n" + logging.RedactSecrets(stderrBuf.String()) + "\n"
	if writeErr := fsutil.SafeWriteFile(logPath, []byte(logPayload), 0o644); writeErr != nil {
		return RunResult{}, writeErr
	}
	if opts.Logger != nil {
		opts.Logger.Verbosef("Ran Codex executable %s with %d args", filepath.Base(exe), len(args))
		if requested := requestedApproval(opts.ApprovalMode); requested != "" && actualApproval != "" && requested != actualApproval {
			opts.Logger.Warnf("Codex CLI reported approval mode %q instead of requested %q. This installed exec implementation may ignore approval overrides in non-interactive mode.", actualApproval, requested)
		}
	}
	if err != nil {
		message := fmt.Sprintf("codex exec failed: %v", err)
		stderr := strings.ToLower(stderrBuf.String())
		if strings.Contains(stderr, "auth") || strings.Contains(stderr, "login") || strings.Contains(stderr, "token") {
			message += ". Run `codex` to complete the official login flow, or follow the Codex CLI API-key login instructions."
		}
		return RunResult{Stdout: stdoutBuf.String(), Stderr: stderrBuf.String(), LogPath: logPath, ActualApprovalMode: actualApproval}, fmt.Errorf("%s", message)
	}
	return RunResult{Stdout: stdoutBuf.String(), Stderr: stderrBuf.String(), LogPath: logPath, ActualApprovalMode: actualApproval}, nil
}

func detectApprovalMode(stderr string) string {
	matches := approvalLinePattern.FindStringSubmatch(stderr)
	if len(matches) != 2 {
		return ""
	}
	return strings.TrimSpace(matches[1])
}

func requestedApproval(value string) string {
	if value == "" {
		return "on-request"
	}
	return value
}
