package spec

import (
	"encoding/json"
	"fmt"

	"github.com/bssm-oss/web-replica/internal/fsutil"
)

func WriteDesignSpec(path string, designSpec DesignSpec) error {
	payload, err := json.MarshalIndent(designSpec, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal design spec: %w", err)
	}
	payload = append(payload, '\n')
	return fsutil.SafeWriteFile(path, payload, 0o644)
}

func WriteJSON(path string, value any) error {
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	payload = append(payload, '\n')
	return fsutil.SafeWriteFile(path, payload, 0o644)
}
