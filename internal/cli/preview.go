package cli

import (
	"context"

	"github.com/bssm-oss/web-replica/internal/logging"
	projectpreview "github.com/bssm-oss/web-replica/internal/preview"
	"github.com/spf13/cobra"
)

func newPreviewCmd(opts *Options) *cobra.Command {
	var port int
	cmd := &cobra.Command{
		Use:   "preview <path>",
		Short: "Preview a generated project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), opts.Timeout)
			defer cancel()
			logger := logging.New(opts.Verbose)
			return projectpreview.RunPreview(ctx, args[0], port, logger)
		},
	}
	cmd.Flags().IntVar(&port, "port", 5173, "Port for preview command")
	return cmd
}
