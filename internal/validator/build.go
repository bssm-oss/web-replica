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
	for _, args := range [][]string{{"npm", installCmd}, {"npm", "run", "build"}} {
		cmd := exec.CommandContext(ctx, args[0], args[1:]...)
		cmd.Dir = projectDir
		cmd.Stdout = &combined
		cmd.Stderr = &combined
		if logger != nil {
			logger.Verbosef("Running %s in %s", args, projectDir)
		}
		if err := cmd.Run(); err != nil {
			logPath := filepath.Join(runDir, "build.log")
			_ = fsutil.SafeWriteFile(logPath, []byte(logging.RedactSecrets(combined.String())), 0o644)
			return BuildResult{LogPath: logPath, Output: combined.String()}, fmt.Errorf("%s failed: %w", args, err)
		}
	}
	logPath := filepath.Join(runDir, "build.log")
	if err := fsutil.SafeWriteFile(logPath, []byte(logging.RedactSecrets(combined.String())), 0o644); err != nil {
		return BuildResult{}, err
	}
	return BuildResult{LogPath: logPath, Output: combined.String()}, nil
}
