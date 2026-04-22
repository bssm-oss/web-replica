package fsutil

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type RunLayout struct {
	OutputRoot     string
	MetadataRoot   string
	RunsRoot       string
	RunID          string
	RunDir         string
	ScreenshotsDir string
	PromptsDir     string
	LatestFile     string
}

func EnsureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}

func SafeJoin(base string, elems ...string) (string, error) {
	joined := filepath.Join(append([]string{base}, elems...)...)
	cleanBase, err := filepath.Abs(base)
	if err != nil {
		return "", fmt.Errorf("resolve base path: %w", err)
	}
	cleanJoined, err := filepath.Abs(joined)
	if err != nil {
		return "", fmt.Errorf("resolve joined path: %w", err)
	}
	if cleanJoined != cleanBase && !strings.HasPrefix(cleanJoined, cleanBase+string(filepath.Separator)) {
		return "", errors.New("path escapes base directory")
	}
	return cleanJoined, nil
}

func NewRunLayout(outputRoot string) (RunLayout, error) {
	runID := time.Now().UTC().Format("2006-01-02T150405Z")
	metadataRoot, err := SafeJoin(outputRoot, ".siteforge")
	if err != nil {
		return RunLayout{}, err
	}
	runsRoot, err := SafeJoin(metadataRoot, "runs")
	if err != nil {
		return RunLayout{}, err
	}
	runDir, err := SafeJoin(runsRoot, runID)
	if err != nil {
		return RunLayout{}, err
	}
	screenshotsDir, err := SafeJoin(runDir, "screenshots")
	if err != nil {
		return RunLayout{}, err
	}
	promptsDir, err := SafeJoin(runDir, "prompts")
	if err != nil {
		return RunLayout{}, err
	}
	latestFile, err := SafeJoin(metadataRoot, "latest.txt")
	if err != nil {
		return RunLayout{}, err
	}
	return RunLayout{
		OutputRoot: outputRoot, MetadataRoot: metadataRoot, RunsRoot: runsRoot, RunID: runID,
		RunDir: runDir, ScreenshotsDir: screenshotsDir, PromptsDir: promptsDir, LatestFile: latestFile,
	}, nil
}

func (r RunLayout) Ensure() error {
	for _, path := range []string{r.OutputRoot, r.MetadataRoot, r.RunsRoot, r.RunDir, r.ScreenshotsDir, r.PromptsDir} {
		if err := EnsureDir(path); err != nil {
			return err
		}
	}
	return os.WriteFile(r.LatestFile, []byte(r.RunID+"\n"), 0o644)
}
