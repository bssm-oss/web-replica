package cli

import (
	"context"
	"fmt"

	"github.com/bssm-oss/web-replica/internal/analyzer"
	"github.com/bssm-oss/web-replica/internal/logging"
	"github.com/spf13/cobra"
)

func newAnalyzeCmd(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "analyze <url>",
		Short: "Analyze a website and write Siteforge artifacts",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), opts.Timeout)
			defer cancel()
			logger := logging.New(opts.Verbose)
			result, err := analyzer.Run(ctx, analyzer.Config{
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
			logger.Infof("design-spec.json: %s", result.DesignSpecPath)
			logger.Infof("brief.md: %s", result.BriefPath)
			logger.Infof("analyzer.log: %s", result.AnalyzerLogPath)
			logger.Infof("screenshots: %s", result.ScreenshotsDir)
			fmt.Fprintf(cmd.OutOrStdout(), "Siteforge analysis complete: %s\n", result.RunDir)
			return nil
		},
	}
}
