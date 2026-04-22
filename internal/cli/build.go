package cli

import (
	"context"

	"github.com/bssm-oss/web-replica/internal/analyzer"
	"github.com/bssm-oss/web-replica/internal/generator"
	"github.com/bssm-oss/web-replica/internal/logging"
	"github.com/bssm-oss/web-replica/internal/validator"
	"github.com/spf13/cobra"
)

func newBuildCmd(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "build <url>",
		Short: "Run the full analyze and generate pipeline",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), opts.Timeout)
			defer cancel()
			logger := logging.New(opts.Verbose)
			analysis, err := analyzer.Run(ctx, analyzer.Config{
				URL:              args[0],
				OutputDir:        opts.OutDir,
				ViewportSelector: opts.Viewport,
				AllowOwnedAssets: opts.AllowOwnedAssets,
				Timeout:          opts.Timeout,
				Logger:           logger,
			})
			if err != nil {
				return err
			}
			if opts.NoAI {
				logger.Infof("Skipping generation because --no-ai was set")
				return nil
			}
			generation, err := generator.Run(ctx, generator.Config{
				SpecPath:          analysis.DesignSpecPath,
				OutputDir:         opts.OutDir,
				Stack:             opts.Stack,
				CodexModel:        opts.CodexModel,
				CodexApprovalMode: opts.CodexApprovalMode,
				Timeout:           opts.Timeout,
				Logger:            logger,
			})
			if err != nil {
				return err
			}
			validation, err := validator.RunBuildAndRepair(ctx, validator.Config{
				ProjectDir:        generation.ProjectDir,
				RunDir:            analysis.RunDir,
				SourceURL:         args[0],
				DesignSpecPath:    analysis.DesignSpecPath,
				CodexModel:        opts.CodexModel,
				CodexApprovalMode: opts.CodexApprovalMode,
				MaxRepairAttempts: 2,
				Timeout:           opts.Timeout,
				Logger:            logger,
			})
			if err != nil {
				return err
			}
			logger.Infof("Build completed successfully.")
			logger.Infof("Run directory: %s", analysis.RunDir)
			logger.Infof("Generated project: %s", generation.ProjectDir)
			logger.Infof("Build log: %s", validation.Build.LogPath)
			logger.Infof("Next steps:")
			logger.Infof("  cd %s", generation.ProjectDir)
			logger.Infof("  npm run preview")
			return nil
		},
	}
}
