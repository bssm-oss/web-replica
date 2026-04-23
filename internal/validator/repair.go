package validator

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bssm-oss/web-replica/internal/codex"
	"github.com/bssm-oss/web-replica/internal/logging"
	"github.com/bssm-oss/web-replica/internal/spec"
)

type Config struct {
	ProjectDir             string
	RunDir                 string
	SourceURL              string
	DesignSpecPath         string
	CodexModel             string
	CodexApprovalMode      string
	MaxRepairAttempts      int
	OriginalScreenshotsDir string
	Timeout                time.Duration
	Logger                 *logging.Logger
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

func appendVisualComparison(cfg Config, notes []string) []string {
	generatedScreenshotPath := extractGeneratedScreenshotPath(notes)
	if cfg.OriginalScreenshotsDir != "" && generatedScreenshotPath != "" {
		visualNotes, _ := CompareWithOriginal(cfg.OriginalScreenshotsDir, generatedScreenshotPath)
		notes = append(notes, visualNotes...)
	}
	return notes
}

func RunBuildAndRepair(ctx context.Context, cfg Config) (Result, error) {
	attempts := cfg.MaxRepairAttempts
	if attempts <= 0 {
		attempts = 5
	}
	var lastBuild BuildResult
	var lastNotes []string
	for attempt := 0; attempt <= attempts; attempt++ {
		buildResult, err := buildProjectFn(ctx, cfg.ProjectDir, cfg.RunDir, cfg.Logger)
		if err == nil {
			notes, noteErr := captureValidationNotesFn(ctx, cfg.ProjectDir, cfg.RunDir)
			if noteErr != nil {
				return Result{Build: buildResult}, nil
			}
			notes = appendVisualComparison(cfg, notes)
			if needsRepair, reason := notesRequireRepair(notes); !needsRepair {
				return Result{Build: buildResult, Notes: notes}, nil
			} else {
				lastBuild = buildResult
				lastNotes = notes
				if attempt == attempts {
					return Result{Build: buildResult, Notes: notes}, fmt.Errorf("visual validation failed after %d repair attempt(s): %s", attempts, reason)
				}
				repairErr := runRepairFn(ctx, cfg, buildResult.Output, notes)
				if repairErr != nil {
					rebuilt, rebuildErr := buildProjectFn(ctx, cfg.ProjectDir, cfg.RunDir, cfg.Logger)
					if rebuildErr == nil {
						finalNotes, _ := captureValidationNotesFn(ctx, cfg.ProjectDir, cfg.RunDir)
						finalNotes = appendVisualComparison(cfg, finalNotes)
						needsRepairAgain, finalReason := notesRequireRepair(finalNotes)
						if !needsRepairAgain {
							finalNotes = append(finalNotes, fmt.Sprintf("repair command reported an error but the subsequent build passed: %v", repairErr))
							return Result{Build: rebuilt, Notes: finalNotes}, nil
						}
						lastBuild = rebuilt
						lastNotes = finalNotes
						if attempt == attempts {
							return Result{Build: rebuilt, Notes: finalNotes}, fmt.Errorf("visual validation failed after repair error: %s", finalReason)
						}
					}
					if cfg.Logger != nil {
						cfg.Logger.Warnf("Repair attempt %d failed: %v", attempt+1, repairErr)
					}
				}
				continue
			}
		}
		lastBuild = buildResult
		lastNotes, _ = captureValidationNotesFn(ctx, cfg.ProjectDir, cfg.RunDir)
		lastNotes = appendVisualComparison(cfg, lastNotes)
		if attempt == attempts {
			return Result{Build: buildResult, Notes: lastNotes}, err
		}
		repairErr := runRepairFn(ctx, cfg, buildResult.Output, lastNotes)
		if repairErr != nil {
			rebuilt, rebuildErr := buildProjectFn(ctx, cfg.ProjectDir, cfg.RunDir, cfg.Logger)
			if rebuildErr == nil {
				finalNotes, _ := captureValidationNotesFn(ctx, cfg.ProjectDir, cfg.RunDir)
				finalNotes = appendVisualComparison(cfg, finalNotes)
				finalNotes = append(finalNotes, fmt.Sprintf("repair command reported an error but the subsequent build passed: %v", repairErr))
				return Result{Build: rebuilt, Notes: finalNotes}, nil
			}
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

func notesRequireRepair(items []string) (bool, string) {
	problemNotes := []string{}
	for _, item := range items {
		switch {
		case item == "body text missing in generated page":
			problemNotes = append(problemNotes, item)
		case item == "horizontal overflow detected in generated page":
			problemNotes = append(problemNotes, item)
		case item == "blank page detected":
			problemNotes = append(problemNotes, item)
		case strings.HasPrefix(item, "visual comparison failed:"):
			problemNotes = append(problemNotes, item)
		}
	}
	if len(problemNotes) == 0 {
		return false, ""
	}
	return true, joinNotes(problemNotes)
}
