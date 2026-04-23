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
	ownedAssets := copyOwnedAssets(runDir, projectDir, designSpec, cfg.Logger)
	rawHTML := ""
	if designSpec.RawHTMLPath != "" {
		rawHTMLFile := filepath.Join(runDir, designSpec.RawHTMLPath)
		if data, err := os.ReadFile(rawHTMLFile); err == nil {
			const maxHTMLBytes = 120 * 1024
			if len(data) > maxHTMLBytes {
				data = data[:maxHTMLBytes]
			}
			rawHTML = string(data)
		}
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
		RawHTML:          rawHTML,
		OwnedAssets:      ownedAssets,
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

func copyOwnedAssets(runDir, projectDir string, designSpec spec.DesignSpec, logger *logging.Logger) []codex.OwnedAsset {
	var owned []codex.OwnedAsset
	allAssets := append(append([]spec.AssetEntry{}, designSpec.Assets.Images...), designSpec.Assets.Fonts...)
	for _, asset := range allAssets {
		if !asset.Allowed || asset.LocalPath == "" {
			continue
		}
		srcPath, err := fsutil.SafeJoin(runDir, filepath.FromSlash(asset.LocalPath))
		if err != nil {
			continue
		}
		data, err := os.ReadFile(srcPath)
		if err != nil {
			continue
		}
		relFromOwned := strings.TrimPrefix(filepath.ToSlash(asset.LocalPath), "owned-assets/")
		publicPath := "/assets/" + relFromOwned
		dstPath, err := fsutil.SafeJoin(projectDir, filepath.FromSlash("public/assets/"+relFromOwned))
		if err != nil {
			continue
		}
		if err := fsutil.SafeWriteFile(dstPath, data, 0o644); err != nil {
			if logger != nil {
				logger.Verbosef("Failed to copy asset %s: %v", asset.LocalPath, err)
			}
			continue
		}
		owned = append(owned, codex.OwnedAsset{PublicPath: publicPath, OriginalURL: asset.URL})
		if logger != nil {
			logger.Verbosef("Copied owned asset → %s", publicPath)
		}
	}
	return owned
}
