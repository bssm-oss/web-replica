package validator

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/bssm-oss/web-replica/internal/fsutil"
	"github.com/bssm-oss/web-replica/internal/logging"
)

type BuildResult struct {
	LogPath string
	Output  string
}

func BuildProject(ctx context.Context, projectDir string, runDir string, logger *logging.Logger) (BuildResult, error) {
	packageJSON := filepath.Join(projectDir, "package.json")
	if _, err := os.Stat(packageJSON); err != nil {
		return BuildResult{}, fmt.Errorf("generated project missing package.json: %w", err)
	}
	installCmd := "install"
	if _, err := os.Stat(filepath.Join(projectDir, "package-lock.json")); err == nil {
		installCmd = "ci"
	}
	var combined bytes.Buffer
	if err := runCommand(ctx, projectDir, logger, &combined, "npm", installCmd); err != nil {
		if installCmd == "ci" {
			if logger != nil {
				logger.Warnf("npm ci failed; retrying with npm install to refresh package-lock.json")
			}
			if retryErr := runCommand(ctx, projectDir, logger, &combined, "npm", "install"); retryErr != nil {
				logPath := filepath.Join(runDir, "build.log")
				_ = fsutil.SafeWriteFile(logPath, []byte(logging.RedactSecrets(combined.String())), 0o644)
				return BuildResult{LogPath: logPath, Output: combined.String()}, fmt.Errorf("[npm install] failed after npm ci fallback: %w", retryErr)
			}
		} else {
			logPath := filepath.Join(runDir, "build.log")
			_ = fsutil.SafeWriteFile(logPath, []byte(logging.RedactSecrets(combined.String())), 0o644)
			return BuildResult{LogPath: logPath, Output: combined.String()}, fmt.Errorf("[npm %s] failed: %w", installCmd, err)
		}
	}
	if err := runCommand(ctx, projectDir, logger, &combined, "npm", "run", "build"); err != nil {
		logPath := filepath.Join(runDir, "build.log")
		_ = fsutil.SafeWriteFile(logPath, []byte(logging.RedactSecrets(combined.String())), 0o644)
		return BuildResult{LogPath: logPath, Output: combined.String()}, fmt.Errorf("[npm run build] failed: %w", err)
	}
	logPath := filepath.Join(runDir, "build.log")
	if err := fsutil.SafeWriteFile(logPath, []byte(logging.RedactSecrets(combined.String())), 0o644); err != nil {
		return BuildResult{}, err
	}
	return BuildResult{LogPath: logPath, Output: combined.String()}, nil
}

func runCommand(ctx context.Context, projectDir string, logger *logging.Logger, output *bytes.Buffer, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = projectDir
	cmd.Stdout = output
	cmd.Stderr = output
	if logger != nil {
		logger.Verbosef("Running %v in %s", append([]string{name}, args...), projectDir)
	}
	return cmd.Run()
}
