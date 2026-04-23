package codex

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/bssm-oss/web-replica/internal/spec"
)

//go:embed generate_site_default.tmpl
var embeddedGenerateTemplate string

//go:embed repair_site_default.tmpl
var embeddedRepairTemplate string

type PromptData struct {
	SourceURL        string
	Stack            string
	Fidelity         string
	FidelityGuidance string
	DesignSpecJSON   string
	BriefMarkdown    string
	ScreenshotPaths  []string
	BuildLogs        string
	ValidationNotes  string
}

func RenderGeneratePrompt(repoRoot string, data PromptData) (string, error) {
	return renderTemplate(loadTemplate(repoRoot, filepath.Join("prompts", "generate_site.md.tmpl"), embeddedGenerateTemplate), data)
}

func RenderRepairPrompt(repoRoot string, data PromptData) (string, error) {
	return renderTemplate(loadTemplate(repoRoot, filepath.Join("prompts", "repair_site.md.tmpl"), embeddedRepairTemplate), data)
}

func loadTemplate(root string, relative string, fallback string) string {
	path := filepath.Join(root, relative)
	contents, err := os.ReadFile(path)
	if err == nil {
		return string(contents)
	}
	return fallback
}

func renderTemplate(source string, data PromptData) (string, error) {
	tmpl, err := template.New("codex").Parse(source)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func CompactSpec(designSpec spec.DesignSpec) string {
	payload, err := json.MarshalIndent(designSpec, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error":%q}`, err.Error())
	}
	return string(payload)
}
