package validator

import (
	"context"
	"fmt"
	"image"
	_ "image/png"
	"math"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bssm-oss/web-replica/internal/browser"
	"github.com/bssm-oss/web-replica/internal/fsutil"
)

func CaptureValidationNotes(ctx context.Context, projectDir string, runDir string) ([]string, error) {
	serveDir := filepath.Join(projectDir, "dist")
	if _, err := os.Stat(serveDir); err != nil {
		serveDir = projectDir
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	defer listener.Close()
	server := &http.Server{Handler: http.FileServer(http.Dir(serveDir))}
	go func() {
		_ = server.Serve(listener)
	}()
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()
	screenshotPath := filepath.Join(runDir, "generated-preview.png")
	if err := fsutil.EnsureDir(filepath.Dir(screenshotPath)); err != nil {
		return nil, err
	}
	info, err := browser.CaptureValidation(ctx, "http://"+listener.Addr().String(), screenshotPath)
	if err != nil {
		return nil, err
	}
	notes := []string{fmt.Sprintf("generated screenshot saved: %s", info.ScreenshotPath)}
	if info.BodyTextPresent {
		notes = append(notes, "body text detected in generated page")
	} else {
		notes = append(notes, "body text missing in generated page")
	}
	if info.HorizontalOverflow {
		notes = append(notes, "horizontal overflow detected in generated page")
	}
	if info.BlankPage {
		notes = append(notes, "blank page detected")
	}
	notes = append(notes, fmt.Sprintf("generated page height: %d", info.PageHeight))
	return notes, nil
}

// CompareWithOriginal compares the generated screenshot against the original desktop screenshot.
// Returns visual diff notes including similarity percentage and zone-level feedback.
func CompareWithOriginal(originalScreenshotsDir string, generatedScreenshotPath string) ([]string, error) {
	originalPath := filepath.Join(originalScreenshotsDir, "desktop.png")
	if _, err := os.Stat(originalPath); err != nil {
		return nil, nil // original screenshot not available, skip comparison
	}
	if _, err := os.Stat(generatedScreenshotPath); err != nil {
		return nil, nil
	}

	origImg, err := loadImage(originalPath)
	if err != nil {
		return nil, nil
	}
	genImg, err := loadImage(generatedScreenshotPath)
	if err != nil {
		return nil, nil
	}

	origBounds := origImg.Bounds()
	genBounds := genImg.Bounds()

	width := origBounds.Max.X
	if genBounds.Max.X < width {
		width = genBounds.Max.X
	}
	height := origBounds.Max.Y
	if genBounds.Max.Y < height {
		height = genBounds.Max.Y
	}
	// Compare only first 2000px vertically to stay within above-fold area
	if height > 2000 {
		height = 2000
	}
	if width == 0 || height == 0 {
		return nil, nil
	}

	const numZones = 4
	zoneNames := []string{"header/nav", "hero", "main content", "footer"}
	zoneDiff := [numZones]int64{}
	zoneTotal := [numZones]int64{}
	totalPixels := int64(width * height)
	matchingPixels := int64(0)

	for y := 0; y < height; y++ {
		zone := (y * numZones) / height
		if zone >= numZones {
			zone = numZones - 1
		}
		for x := 0; x < width; x++ {
			or_, og, ob, _ := origImg.At(x, y).RGBA()
			gr, gg, gb, _ := genImg.At(x, y).RGBA()
			dr := math.Abs(float64(or_>>8) - float64(gr>>8))
			dg := math.Abs(float64(og>>8) - float64(gg>>8))
			db := math.Abs(float64(ob>>8) - float64(gb>>8))
			diff := (dr + dg + db) / 3.0
			zoneTotal[zone]++
			if diff < 25 {
				matchingPixels++
			} else {
				zoneDiff[zone]++
			}
		}
	}

	similarity := float64(matchingPixels) / float64(totalPixels) * 100
	notes := []string{fmt.Sprintf("visual similarity: %.0f%%", similarity)}

	for i, name := range zoneNames {
		if zoneTotal[i] == 0 {
			continue
		}
		zoneSimilarity := (1.0 - float64(zoneDiff[i])/float64(zoneTotal[i])) * 100
		if zoneSimilarity < 70 {
			notes = append(notes, fmt.Sprintf("%s zone differs significantly: %.0f%% match - needs visual fix", name, zoneSimilarity))
		}
	}

	if similarity >= 40 {
		notes = append(notes, "visual comparison passed")
	} else {
		notes = append(notes, fmt.Sprintf("visual comparison failed: %.0f%% similarity is below 40%% threshold - fix layout, colors, typography, and section structure to match the original (note: CSS-only pages cannot pixel-match photographic originals; focus on structural and color accuracy)", similarity))
	}

	return notes, nil
}

func loadImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	return img, err
}

func extractGeneratedScreenshotPath(notes []string) string {
	for _, note := range notes {
		if strings.HasPrefix(note, "generated screenshot saved: ") {
			return strings.TrimPrefix(note, "generated screenshot saved: ")
		}
	}
	return ""
}
