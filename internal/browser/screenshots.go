package browser

import "path/filepath"

func RelativeScreenshotPath(runDir string, absolutePath string) string {
	rel, err := filepath.Rel(runDir, absolutePath)
	if err != nil {
		return absolutePath
	}
	return filepath.ToSlash(rel)
}
