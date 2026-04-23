package generator

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bssm-oss/web-replica/internal/codex"
	"github.com/bssm-oss/web-replica/internal/fsutil"
	"github.com/bssm-oss/web-replica/internal/logging"
	"github.com/bssm-oss/web-replica/internal/spec"
)

type Config struct {
	SpecPath          string
	OutputDir         string
	Stack             string
	Fidelity          string
	CodexModel        string
	CodexApprovalMode string
	Timeout           time.Duration
	Logger            *logging.Logger
}

type Result struct {
	ProjectDir string
	RunDir     string
	PromptPath string
	LogPath    string
}

func Run(ctx context.Context, cfg Config) (Result, error) {
	if err := ValidateStack(cfg.Stack); err != nil {
		return Result{}, err
	}
	if !IsImplementedStack(cfg.Stack) {
		return Result{}, fmt.Errorf("stack %q is reserved for a future implementation", cfg.Stack)
	}
	fidelity, err := NormalizeFidelity(cfg.Fidelity)
	if err != nil {
		return Result{}, err
	}
	specPayload, err := os.ReadFile(cfg.SpecPath)
	if err != nil {
		return Result{}, err
	}
	var designSpec spec.DesignSpec
	if err := json.Unmarshal(specPayload, &designSpec); err != nil {
		return Result{}, fmt.Errorf("parse design spec: %w", err)
	}
	runDir := filepath.Dir(cfg.SpecPath)
	briefPath := filepath.Join(runDir, "brief.md")
	briefPayload, err := os.ReadFile(briefPath)
	if err != nil {
		return Result{}, err
	}
	projectDir := cfg.OutputDir
	if strings.HasSuffix(projectDir, ".siteforge") {
		projectDir = filepath.Join(filepath.Dir(projectDir), "generated-site")
	}
	if err := fsutil.EnsureDir(projectDir); err != nil {
		return Result{}, err
	}
	repoRoot, _ := os.Getwd()
	prompt, err := codex.RenderGeneratePrompt(repoRoot, codex.PromptData{
		SourceURL:        designSpec.SourceURL,
		Stack:            cfg.Stack,
		Fidelity:         fidelity,
		FidelityGuidance: FidelityGuidance(fidelity),
		DesignSpecJSON:   codex.CompactSpec(designSpec),
		BriefMarkdown:    string(briefPayload),
		ScreenshotPaths:  collectScreenshotPaths(runDir, designSpec),
	})
	if err != nil {
		return Result{}, err
	}
	promptPath := filepath.Join(runDir, "prompts", "generate_site.md")
	result, err := codex.Run(ctx, codex.RunOptions{OutputDir: projectDir, RunDir: runDir, Prompt: prompt, PromptPath: promptPath, ApprovalMode: cfg.CodexApprovalMode, Model: cfg.CodexModel, Logger: cfg.Logger})
	if err != nil {
		return Result{}, err
	}
	return Result{ProjectDir: projectDir, RunDir: runDir, PromptPath: promptPath, LogPath: result.LogPath}, nil
}

func collectScreenshotPaths(runDir string, designSpec spec.DesignSpec) []string {
	paths := []string{}
	for _, value := range []string{designSpec.Responsive.Desktop.Screenshot, designSpec.Responsive.Tablet.Screenshot, designSpec.Responsive.Mobile.Screenshot} {
		if value != "" {
			paths = append(paths, filepath.Join(runDir, filepath.FromSlash(value)))
		}
	}
	return paths
}
