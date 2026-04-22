package validator

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/bssm-oss/web-replica/internal/codex"
	"github.com/bssm-oss/web-replica/internal/logging"
	"github.com/bssm-oss/web-replica/internal/spec"
)

type Config struct {
	ProjectDir        string
	RunDir            string
	SourceURL         string
	DesignSpecPath    string
	CodexModel        string
	CodexApprovalMode string
	Timeout           time.Duration
	Logger            *logging.Logger
}

type Result struct {
	Build BuildResult
	Notes []string
}

func RunBuildAndRepair(ctx context.Context, cfg Config) (Result, error) {
	buildResult, err := BuildProject(ctx, cfg.ProjectDir, cfg.RunDir, cfg.Logger)
	if err == nil {
		notes, noteErr := CaptureValidationNotes(ctx, cfg.ProjectDir, cfg.RunDir)
		if noteErr == nil {
			return Result{Build: buildResult, Notes: notes}, nil
		}
		return Result{Build: buildResult}, nil
	}
	notes, _ := CaptureValidationNotes(ctx, cfg.ProjectDir, cfg.RunDir)
	repairErr := runRepair(ctx, cfg, buildResult.Output, notes)
	if repairErr != nil {
		rebuilt, rebuildErr := BuildProject(ctx, cfg.ProjectDir, cfg.RunDir, cfg.Logger)
		if rebuildErr == nil {
			finalNotes, _ := CaptureValidationNotes(ctx, cfg.ProjectDir, cfg.RunDir)
			finalNotes = append(finalNotes, fmt.Sprintf("repair command reported an error but build passed afterward: %v", repairErr))
			return Result{Build: rebuilt, Notes: finalNotes}, nil
		}
		return Result{Build: buildResult, Notes: notes}, fmt.Errorf("initial build failed and repair failed: %w", repairErr)
	}
	rebuilt, rebuildErr := BuildProject(ctx, cfg.ProjectDir, cfg.RunDir, cfg.Logger)
	if rebuildErr != nil {
		return Result{Build: rebuilt, Notes: notes}, rebuildErr
	}
	finalNotes, _ := CaptureValidationNotes(ctx, cfg.ProjectDir, cfg.RunDir)
	return Result{Build: rebuilt, Notes: finalNotes}, nil
}

func runRepair(ctx context.Context, cfg Config, buildLogs string, validationNotes []string) error {
	payload, err := os.ReadFile(cfg.DesignSpecPath)
	if err != nil {
		return err
	}
	var designSpec spec.DesignSpec
	if err := json.Unmarshal(payload, &designSpec); err != nil {
		return err
	}
	repoRoot, _ := os.Getwd()
	prompt, err := codex.RenderRepairPrompt(repoRoot, codex.PromptData{
		SourceURL:       cfg.SourceURL,
		DesignSpecJSON:  codex.CompactSpec(designSpec),
		BuildLogs:       buildLogs,
		ValidationNotes: joinNotes(validationNotes),
	})
	if err != nil {
		return err
	}
	promptPath := filepath.Join(cfg.RunDir, "prompts", "repair_site.md")
	_, err = codex.Run(ctx, codex.RunOptions{OutputDir: cfg.ProjectDir, RunDir: cfg.RunDir, Prompt: prompt, PromptPath: promptPath, ApprovalMode: cfg.CodexApprovalMode, Model: cfg.CodexModel, Logger: cfg.Logger})
	return err
}

func joinNotes(items []string) string {
	if len(items) == 0 {
		return "no validation notes collected"
	}
	value := ""
	for _, item := range items {
		value += "- " + item + "\n"
	}
	return value
}
