package validator

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
	cmd.Env = nodeEnv()
	cmd.Stdout = output
	cmd.Stderr = output
	if logger != nil {
		logger.Verbosef("Running %v in %s", append([]string{name}, args...), projectDir)
	}
	return cmd.Run()
}

// nodeEnv returns the current environment augmented with paths where Node/npm
// are commonly installed (nvm, Homebrew, system Node, volta, fnm, etc.) so that
// builds succeed even when webreplica is invoked from a context with a minimal PATH.
func nodeEnv() []string {
	env := os.Environ()
	extraPaths := []string{
		"/usr/local/bin",
		"/usr/bin",
		"/bin",
		"/opt/homebrew/bin",
		"/opt/homebrew/sbin",
		"/usr/local/sbin",
	}
	// Add nvm versions directory if present
	if home, err := os.UserHomeDir(); err == nil {
		nvmBase := filepath.Join(home, ".nvm", "versions", "node")
		if entries, err := os.ReadDir(nvmBase); err == nil {
			// Pick the highest-versioned node directory
			for i := len(entries) - 1; i >= 0; i-- {
				if entries[i].IsDir() {
					extraPaths = append([]string{filepath.Join(nvmBase, entries[i].Name(), "bin")}, extraPaths...)
					break
				}
			}
		}
		// volta
		extraPaths = append(extraPaths, filepath.Join(home, ".volta", "bin"))
		// fnm
		extraPaths = append(extraPaths, filepath.Join(home, ".fnm", "aliases", "default", "bin"))
	}
	// Prepend extra paths to existing PATH
	existingPath := os.Getenv("PATH")
	newPath := strings.Join(extraPaths, string(os.PathListSeparator))
	if existingPath != "" {
		newPath += string(os.PathListSeparator) + existingPath
	}
	result := make([]string, 0, len(env)+1)
	for _, e := range env {
		if !strings.HasPrefix(e, "PATH=") {
			result = append(result, e)
		}
	}
	result = append(result, "PATH="+newPath)
	return result
}
