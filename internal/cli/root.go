package cli

import (
	"fmt"
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

type RootConfig struct {
	Name          string
	DefaultOutDir string
}

func NewRootCmd() *cobra.Command {
	return NewRootCmdWithConfig(RootConfig{Name: "siteforge", DefaultOutDir: "./siteforge-output"})
}

func NewWebReplicaCmd() *cobra.Command {
	return NewRootCmdWithConfig(RootConfig{Name: "webreplica", DefaultOutDir: "./generated-site"})
}

func NewRootCmdWithConfig(cfg RootConfig) *cobra.Command {
	if cfg.Name == "" {
		cfg.Name = "siteforge"
	}
	if cfg.DefaultOutDir == "" {
		cfg.DefaultOutDir = "./siteforge-output"
	}
	opts := &Options{}
	cmd := &cobra.Command{
		Use:           fmt.Sprintf("%s [url]", cfg.Name),
		Short:         "Generate safe frontend reimplementations from authorized website URLs",
		Long:          "Analyze a publicly accessible website and generate an inspired, safety-aware frontend reimplementation without copying protected branding, long-form text, or tracking code.\n\nQuick use:\n  " + cfg.Name + " https://example.com",
		Args:          cobra.MaximumNArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			return runBuild(cmd, opts, args[0])
		},
	}
	cmd.PersistentFlags().StringVar(&opts.OutDir, "out", cfg.DefaultOutDir, "Output directory for generated artifacts or projects")
	cmd.PersistentFlags().StringVar(&opts.Stack, "stack", "vite-react-tailwind", "Frontend stack to generate")
	cmd.PersistentFlags().StringVar(&opts.Viewport, "viewport", "desktop,tablet,mobile", "Viewport selection: desktop, tablet, mobile, all, or comma-separated")
	cmd.PersistentFlags().BoolVar(&opts.AllowOwnedAssets, "allow-owned-assets", false, "Allow same-origin owned image/font downloads")
	cmd.PersistentFlags().BoolVar(&opts.NoAI, "no-ai", false, "Skip Codex generation and only create analysis artifacts")
	cmd.PersistentFlags().BoolVar(&opts.Verbose, "verbose", false, "Enable verbose logs")
	cmd.PersistentFlags().StringVar(&opts.CodexModel, "codex-model", "", "Optional Codex model override")
	cmd.PersistentFlags().StringVar(&opts.CodexApprovalMode, "codex-approval-mode", "on-request", "Codex approval mode: on-request, untrusted, never")
	cmd.PersistentFlags().BoolVar(&opts.KeepWorkdir, "keep-workdir", false, "Keep temporary work directories after completion")
	cmd.PersistentFlags().DurationVar(&opts.Timeout, "timeout", 20*time.Minute, "Timeout for network, browser, and Codex operations")

	cmd.AddCommand(newDoctorCmd(opts))
	cmd.AddCommand(newAnalyzeCmd(opts))
	cmd.AddCommand(newGenerateCmd(opts))
	cmd.AddCommand(newBuildCmd(opts))
	cmd.AddCommand(newPreviewCmd(opts))
	return cmd
}
