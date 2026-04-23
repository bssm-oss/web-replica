package cli

import (
	"context"

	"github.com/bssm-oss/web-replica/internal/codex"
	"github.com/spf13/cobra"
)

func newLoginCmd(opts *Options) *cobra.Command {
	var status bool
	var deviceAuth bool
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Open the official Codex login flow",
		Long:  "Runs the official `codex login` command. Webreplica does not implement OAuth directly and never reads Codex token files.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), opts.Timeout)
			defer cancel()
			return codex.RunLogin(ctx, codex.LoginOptions{Status: status, DeviceAuth: deviceAuth})
		},
	}
	cmd.Flags().BoolVar(&status, "status", false, "Show official Codex login status")
	cmd.Flags().BoolVar(&deviceAuth, "device-auth", false, "Use Codex device auth flow")
	return cmd
}
