package cli

import (
	"time"

	"github.com/spf13/cobra"
)

type Options struct {
	OutDir            string
	Stack             string
	Viewport          string
	AllowOwnedAssets  bool
	NoAI              bool
	Verbose           bool
	CodexModel        string
	CodexApprovalMode string
	KeepWorkdir       bool
	Timeout           time.Duration
}

func NewRootCmd() *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:           "siteforge",
		Short:         "Generate safe frontend reimplementations from authorized website URLs",
		Long:          "Siteforge analyzes publicly accessible websites and generates inspired, safety-aware frontend reimplementations without copying protected branding, long-form text, or tracking code.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.PersistentFlags().StringVar(&opts.OutDir, "out", "./siteforge-output", "Output directory for generated artifacts or projects")
	cmd.PersistentFlags().StringVar(&opts.Stack, "stack", "vite-react-tailwind", "Frontend stack to generate")
	cmd.PersistentFlags().StringVar(&opts.Viewport, "viewport", "desktop,tablet,mobile", "Viewport selection: desktop, tablet, mobile, all, or comma-separated")
	cmd.PersistentFlags().BoolVar(&opts.AllowOwnedAssets, "allow-owned-assets", false, "Allow same-origin owned image/font downloads")
	cmd.PersistentFlags().BoolVar(&opts.NoAI, "no-ai", false, "Skip Codex generation and only create analysis artifacts")
	cmd.PersistentFlags().BoolVar(&opts.Verbose, "verbose", false, "Enable verbose logs")
	cmd.PersistentFlags().StringVar(&opts.CodexModel, "codex-model", "", "Optional Codex model override")
	cmd.PersistentFlags().StringVar(&opts.CodexApprovalMode, "codex-approval-mode", "on-request", "Codex approval mode: on-request, untrusted, never")
	cmd.PersistentFlags().BoolVar(&opts.KeepWorkdir, "keep-workdir", false, "Keep temporary work directories after completion")
	cmd.PersistentFlags().DurationVar(&opts.Timeout, "timeout", 2*time.Minute, "Timeout for network, browser, and Codex operations")

	cmd.AddCommand(newDoctorCmd(opts))
	cmd.AddCommand(newAnalyzeCmd(opts))
	cmd.AddCommand(newGenerateCmd(opts))
	cmd.AddCommand(newBuildCmd(opts))
	cmd.AddCommand(newPreviewCmd(opts))
	return cmd
}
