package fsutil

import (
	"fmt"
	"os"
	"path/filepath"
)

func SafeWriteFile(path string, data []byte, perm os.FileMode) error {
	if err := EnsureDir(filepath.Dir(path)); err != nil {
		return fmt.Errorf("ensure parent directory: %w", err)
	}
	return os.WriteFile(path, data, perm)
}

func SafeWriteRelative(base string, relative string, data []byte, perm os.FileMode) (string, error) {
	resolved, err := SafeJoin(base, relative)
	if err != nil {
		return "", err
	}
	if err := SafeWriteFile(resolved, data, perm); err != nil {
		return "", err
	}
	return resolved, nil
}
