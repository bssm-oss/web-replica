package validator

import (
	"context"
	"errors"
	"testing"

	"github.com/bssm-oss/web-replica/internal/logging"
)

func TestRunBuildAndRepairRetriesAndSucceeds(t *testing.T) {
	originalBuild := buildProjectFn
	originalCapture := captureValidationNotesFn
	originalRepair := runRepairFn
	defer func() {
		buildProjectFn = originalBuild
		captureValidationNotesFn = originalCapture
		runRepairFn = originalRepair
	}()
	buildCalls := 0
	repairCalls := 0
	buildProjectFn = func(context.Context, string, string, *logging.Logger) (BuildResult, error) {
		buildCalls++
		if buildCalls == 1 {
			return BuildResult{LogPath: "build.log", Output: "failed"}, errors.New("boom")
		}
		return BuildResult{LogPath: "build.log", Output: "ok"}, nil
	}
	captureValidationNotesFn = func(context.Context, string, string) ([]string, error) {
		return []string{"note"}, nil
	}
	runRepairFn = func(context.Context, Config, string, []string) error {
		repairCalls++
		return nil
	}
	result, err := RunBuildAndRepair(context.Background(), Config{MaxRepairAttempts: 2})
	if err != nil {
		t.Fatalf("RunBuildAndRepair returned error: %v", err)
	}
	if buildCalls != 2 {
		t.Fatalf("expected 2 build attempts, got %d", buildCalls)
	}
	if repairCalls != 1 {
		t.Fatalf("expected 1 repair attempt, got %d", repairCalls)
	}
	if result.Build.Output != "ok" {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestRunBuildAndRepairStopsAfterMaxAttempts(t *testing.T) {
	originalBuild := buildProjectFn
	originalCapture := captureValidationNotesFn
	originalRepair := runRepairFn
	defer func() {
		buildProjectFn = originalBuild
		captureValidationNotesFn = originalCapture
		runRepairFn = originalRepair
	}()
	buildProjectFn = func(context.Context, string, string, *logging.Logger) (BuildResult, error) {
		return BuildResult{LogPath: "build.log", Output: "failed"}, errors.New("boom")
	}
	captureValidationNotesFn = func(context.Context, string, string) ([]string, error) {
		return []string{"note"}, nil
	}
	repairCalls := 0
	runRepairFn = func(context.Context, Config, string, []string) error {
		repairCalls++
		return errors.New("repair failed")
	}
	_, err := RunBuildAndRepair(context.Background(), Config{MaxRepairAttempts: 2})
	if err == nil {
		t.Fatal("expected failure after exhausting repair attempts")
	}
	if repairCalls != 2 {
		t.Fatalf("expected 2 repair attempts, got %d", repairCalls)
	}
}

func TestNotesRequireRepair(t *testing.T) {
	needsRepair, reason := notesRequireRepair([]string{
		"generated screenshot saved: test.png",
		"body text missing in generated page",
		"blank page detected",
	})
	if !needsRepair {
		t.Fatal("expected notes to require repair")
	}
	if reason == "" {
		t.Fatal("expected repair reason to be populated")
	}
}

func TestNotesRequireRepairIgnoresHealthyNotes(t *testing.T) {
	needsRepair, reason := notesRequireRepair([]string{
		"generated screenshot saved: test.png",
		"body text detected in generated page",
		"generated page height: 1200",
	})
	if needsRepair {
		t.Fatalf("expected notes to be considered healthy, got %q", reason)
	}
}
