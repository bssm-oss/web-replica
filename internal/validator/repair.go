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
	MaxRepairAttempts int
	Timeout           time.Duration
	Logger            *logging.Logger
}

type Result struct {
	Build BuildResult
	Notes []string
}

var (
	buildProjectFn           = BuildProject
	captureValidationNotesFn = CaptureValidationNotes
	runRepairFn              = runRepair
)

func RunBuildAndRepair(ctx context.Context, cfg Config) (Result, error) {
	attempts := cfg.MaxRepairAttempts
	if attempts <= 0 {
		attempts = 2
	}
	var lastBuild BuildResult
	var lastNotes []string
	for attempt := 0; attempt <= attempts; attempt++ {
		buildResult, err := buildProjectFn(ctx, cfg.ProjectDir, cfg.RunDir, cfg.Logger)
		if err == nil {
			notes, noteErr := captureValidationNotesFn(ctx, cfg.ProjectDir, cfg.RunDir)
			if noteErr == nil {
				return Result{Build: buildResult, Notes: notes}, nil
			}
			return Result{Build: buildResult}, nil
		}
		lastBuild = buildResult
		lastNotes, _ = captureValidationNotesFn(ctx, cfg.ProjectDir, cfg.RunDir)
		if attempt == attempts {
			return Result{Build: buildResult, Notes: lastNotes}, err
		}
		repairErr := runRepairFn(ctx, cfg, buildResult.Output, lastNotes)
		if repairErr != nil {
			if cfg.Logger != nil {
				cfg.Logger.Warnf("Repair attempt %d failed: %v", attempt+1, repairErr)
			}
			continue
		}
	}
	return Result{Build: lastBuild, Notes: lastNotes}, fmt.Errorf("build failed after %d repair attempt(s)", attempts)
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
