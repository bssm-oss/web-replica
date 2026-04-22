package analyzer

import (
	"context"
	"path/filepath"
	"strconv"
	"time"

	"github.com/bssm-oss/web-replica/internal/browser"
	"github.com/bssm-oss/web-replica/internal/fsutil"
	"github.com/bssm-oss/web-replica/internal/logging"
	"github.com/bssm-oss/web-replica/internal/spec"
)

type Config struct {
	URL              string
	OutputDir        string
	ViewportSelector string
	AllowOwnedAssets bool
	Timeout          time.Duration
	Logger           *logging.Logger
}

type Result struct {
	RunDir          string
	DesignSpecPath  string
	BriefPath       string
	RawOutlinePath  string
	ScreenshotsDir  string
	AnalyzerLogPath string
	DesignSpec      spec.DesignSpec
}

func Run(ctx context.Context, cfg Config) (Result, error) {
	validated, err := ValidateURL(ctx, cfg.URL)
	if err != nil {
		return Result{}, err
	}
	layout, err := fsutil.NewRunLayout(cfg.OutputDir)
	if err != nil {
		return Result{}, err
	}
	if err := layout.Ensure(); err != nil {
		return Result{}, err
	}
	fetched, err := FetchHTML(ctx, validated, cfg.Timeout)
	if err != nil {
		return Result{}, err
	}
	htmlAnalysis, err := AnalyzeHTML(fetched.Body, validated.Normalized, cfg.AllowOwnedAssets)
	if err != nil {
		return Result{}, err
	}
	if cfg.AllowOwnedAssets {
		htmlAnalysis.CandidateAssets = DownloadOwnedAssets(ctx, htmlAnalysis.CandidateAssets, layout.RunDir, cfg.Logger)
	}
	viewports, err := browser.ResolveViewports(cfg.ViewportSelector)
	if err != nil {
		return Result{}, err
	}
	captures, err := browser.CapturePage(ctx, validated.Normalized, layout.ScreenshotsDir, viewports)
	if err != nil {
		return Result{}, err
	}
	designSpec := buildDesignSpec(validated, fetched, htmlAnalysis, captures, layout.RunDir, cfg.AllowOwnedAssets)
	designSpecPath := filepath.Join(layout.RunDir, "design-spec.json")
	briefPath := filepath.Join(layout.RunDir, "brief.md")
	rawOutlinePath := filepath.Join(layout.RunDir, "raw-outline.json")
	analyzerLogPath := filepath.Join(layout.RunDir, "analyzer.log")
	if err := spec.WriteDesignSpec(designSpecPath, designSpec); err != nil {
		return Result{}, err
	}
	if err := spec.WriteBrief(briefPath, designSpec); err != nil {
		return Result{}, err
	}
	if err := spec.WriteJSON(rawOutlinePath, htmlAnalysis); err != nil {
		return Result{}, err
	}
	if err := fsutil.SafeWriteFile(analyzerLogPath, []byte(buildAnalyzerLog(validated, fetched, htmlAnalysis, captures)), 0o644); err != nil {
		return Result{}, err
	}
	if cfg.Logger != nil {
		cfg.Logger.Verbosef("Analysis artifacts written to %s", layout.RunDir)
	}
	return Result{RunDir: layout.RunDir, DesignSpecPath: designSpecPath, BriefPath: briefPath, RawOutlinePath: rawOutlinePath, ScreenshotsDir: layout.ScreenshotsDir, AnalyzerLogPath: analyzerLogPath, DesignSpec: designSpec}, nil
}

func buildDesignSpec(validated ValidatedURL, fetched FetchedPage, htmlAnalysis HTMLAnalysis, captures []browser.ViewportCapture, runDir string, allowOwnedAssets bool) spec.DesignSpec {
	responsive := spec.ResponsiveSpec{}
	for _, capture := range captures {
		viewport := spec.ViewportAnalysis{Screenshot: browser.RelativeScreenshotPath(runDir, capture.ScreenshotPath), Notes: capture.Notes}
		switch capture.Viewport.Name {
		case "desktop":
			responsive.Desktop = viewport
		case "tablet":
			responsive.Tablet = viewport
		case "mobile":
			responsive.Mobile = viewport
		}
	}
	tokens := browser.BuildDesignTokens(captures)
	return spec.DesignSpec{
		SchemaVersion: "0.1",
		SourceURL:     validated.Source,
		NormalizedURL: validated.Normalized,
		Mode:          "inspired_reimplementation",
		CreatedAt:     fetched.FetchedAt.Format(time.RFC3339),
		Page: spec.Page{
			Title:          htmlAnalysis.Title,
			Description:    htmlAnalysis.Description,
			Language:       htmlAnalysis.Language,
			ContentSummary: htmlAnalysis.ContentSummary,
			Structure:      htmlAnalysis.Structure,
		},
		DesignTokens: tokens,
		Responsive:   responsive,
		Assets: spec.AssetPolicy{
			Policy:             "placeholders_by_default",
			AllowedOwnedAssets: allowOwnedAssets,
			Images:             filterAssets(htmlAnalysis.CandidateAssets, "image", allowOwnedAssets),
			Fonts:              filterAssets(htmlAnalysis.CandidateAssets, "font", allowOwnedAssets),
		},
		GenerationRules: []string{
			"Do not copy protected logos or branding.",
			"Do not copy long original text.",
			"Create original placeholder copy.",
			"Use responsive accessible components.",
			"Do not include third-party tracking scripts.",
		},
	}
}

func filterAssets(assets []spec.AssetEntry, mimeType string, allowOwnedAssets bool) []spec.AssetEntry {
	items := []spec.AssetEntry{}
	for _, asset := range assets {
		if asset.MimeType == mimeType {
			if !allowOwnedAssets && asset.Allowed {
				asset.Allowed = false
				asset.LocalPath = ""
				asset.Reason = "asset downloads disabled without --allow-owned-assets"
			}
			items = append(items, asset)
		}
	}
	return items
}

func buildAnalyzerLog(validated ValidatedURL, fetched FetchedPage, htmlAnalysis HTMLAnalysis, captures []browser.ViewportCapture) string {
	log := "Siteforge Analyzer\n\n"
	log += "Source URL: " + validated.Source + "\n"
	log += "Normalized URL: " + validated.Normalized + "\n"
	log += "Hostname: " + validated.Hostname + "\n"
	log += "Scheme: " + validated.Scheme + "\n"
	log += "Fetched At: " + fetched.FetchedAt.Format(time.RFC3339) + "\n"
	log += "Content Type: " + fetched.ContentType + "\n"
	log += "Headings: " + strconv.Itoa(len(htmlAnalysis.Structure.Headings)) + "\n"
	log += "Navigation Items: " + strconv.Itoa(len(htmlAnalysis.Structure.Navigation)) + "\n"
	log += "Sections: " + strconv.Itoa(len(htmlAnalysis.Structure.Sections)) + "\n"
	log += "Forms: " + strconv.Itoa(len(htmlAnalysis.Structure.Forms)) + "\n"
	log += "Links: " + strconv.Itoa(len(htmlAnalysis.Structure.Links)) + "\n"
	log += "Images: " + strconv.Itoa(len(htmlAnalysis.Structure.Images)) + "\n"
	for _, capture := range captures {
		log += "Viewport " + capture.Viewport.Name + ": " + capture.ScreenshotPath + "\n"
	}
	return log
}
