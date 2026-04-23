package cli

import (
	"context"
	"errors"

	"github.com/bssm-oss/web-replica/internal/generator"
	"github.com/bssm-oss/web-replica/internal/logging"
	"github.com/spf13/cobra"
)

func newGenerateCmd(opts *Options) *cobra.Command {
	var specPath string
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate a frontend project from a design spec",
		RunE: func(cmd *cobra.Command, args []string) error {
			if specPath == "" {
				return errors.New("--spec is required")
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), opts.Timeout)
			defer cancel()
			logger := logging.New(opts.Verbose)
			_, err := generator.Run(ctx, generator.Config{
				SpecPath:          specPath,
				OutputDir:         opts.OutDir,
				Stack:             opts.Stack,
				Fidelity:          opts.Fidelity,
				CodexModel:        opts.CodexModel,
				CodexApprovalMode: opts.CodexApprovalMode,
				Timeout:           opts.Timeout,
				Logger:            logger,
			})
			return err
		},
	}
	cmd.Flags().StringVar(&specPath, "spec", "", "Path to design-spec.json")
	return cmd
}
