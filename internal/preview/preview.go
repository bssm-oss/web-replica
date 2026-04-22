package preview

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/bssm-oss/web-replica/internal/logging"
)

func RunDoctor(ctx context.Context, logger *logging.Logger) error {
	checks := []struct {
		Label string
		Cmd   []string
		Hint  string
	}{
		{Label: "Go", Cmd: []string{"go", "version"}},
		{Label: "Git", Cmd: []string{"git", "version"}},
		{Label: "Node", Cmd: []string{"node", "--version"}},
		{Label: "npm", Cmd: []string{"npm", "--version"}},
		{Label: "Codex", Cmd: []string{"codex", "--version"}, Hint: "Install with:\n  npm i -g @openai/codex\n  codex"},
	}
	logger.Infof("Siteforge Doctor")
	logger.Infof("")
	var firstErr error
	for _, check := range checks {
		output, err := runProbe(ctx, check.Cmd[0], check.Cmd[1:]...)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			logger.Warnf("%s:\tMISSING\t%s", check.Label, check.Hint)
			continue
		}
		logger.Infof("%s:\tOK\t%s", check.Label, output)
	}
	chrome, err := findChrome()
	if err != nil {
		if firstErr == nil {
			firstErr = err
		}
		logger.Warnf("Chrome:\tMISSING\tInstall Chrome or Chromium for screenshot capture")
	} else {
		logger.Infof("Chrome:\tOK\t%s", chrome)
	}
	if firstErr == nil {
		logger.Infof("\nAll required tools are available.")
	}
	return firstErr
}

func RunPreview(ctx context.Context, targetPath string, port int, logger *logging.Logger) error {
	packageJSON := filepath.Join(targetPath, "package.json")
	if _, err := os.Stat(packageJSON); err == nil {
		for _, args := range previewCommands(port) {
			cmd := exec.CommandContext(ctx, args[0], args[1:]...)
			cmd.Dir = targetPath
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err == nil {
				return nil
			}
		}
	}
	distDir := filepath.Join(targetPath, "dist")
	if _, err := os.Stat(distDir); err == nil {
		logger.Infof("Static build available at %s", distDir)
		logger.Infof("Open index.html directly or run: python3 -m http.server %d --directory %s", port, distDir)
		return nil
	}
	return fmt.Errorf("no previewable project found in %s", targetPath)
}

func previewCommands(port int) [][]string {
	return [][]string{{"npm", "run", "preview", "--", "--port", fmt.Sprintf("%d", port)}, {"npm", "run", "dev", "--", "--port", fmt.Sprintf("%d", port)}}
}

func runProbe(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(bytesTrim(output)), nil
}

func bytesTrim(input []byte) []byte {
	for len(input) > 0 && (input[len(input)-1] == '\n' || input[len(input)-1] == '\r' || input[len(input)-1] == '\t' || input[len(input)-1] == ' ') {
		input = input[:len(input)-1]
	}
	return input
}

func findChrome() (string, error) {
	for _, name := range []string{"google-chrome", "chromium", "chromium-browser", "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"} {
		if filepath.IsAbs(name) {
			if _, err := os.Stat(name); err == nil {
				return name, nil
			}
			continue
		}
		if path, err := exec.LookPath(name); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("chrome not found")
}
