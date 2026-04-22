package cli

import (
	"context"

	"github.com/bssm-oss/web-replica/internal/logging"
	"github.com/bssm-oss/web-replica/internal/preview"
	"github.com/spf13/cobra"
)

func newDoctorCmd(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check Siteforge prerequisites",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), opts.Timeout)
			defer cancel()
			logger := logging.New(opts.Verbose)
			return preview.RunDoctor(ctx, logger)
		},
	}
}
