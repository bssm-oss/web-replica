package validator

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
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
