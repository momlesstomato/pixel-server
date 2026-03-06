package cli

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"pixelsv/pkg/config"
	logpkg "pixelsv/pkg/log"
)

// newServeCommand creates the runtime serve command.
func newServeCommand() *cobra.Command {
	var envFile string
	var role string
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "start core HTTP/WebSocket runtime",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			v, err := config.NewViper(config.LoadOptions{EnvFile: envFile})
			if err != nil {
				return err
			}
			if cmd.Flags().Changed("role") {
				v.Set("runtime.role", role)
			}
			if err := logpkg.BindViper(v); err != nil {
				return err
			}
			baseCfg, err := config.FromViper(v)
			if err != nil {
				return err
			}
			roles, err := newRoleSet(baseCfg.Runtime.Role)
			if err != nil {
				return err
			}
			logCfg, err := logpkg.FromViper(v)
			if err != nil {
				return err
			}
			logger, err := logpkg.New(logCfg)
			if err != nil {
				return err
			}
			defer logger.Sync()
			return runRoleAwareStartup(ctx, v, logger, roles)
		},
	}
	cmd.Flags().StringVar(&envFile, "env-file", ".env", "env file path")
	cmd.Flags().StringVar(&role, "role", "all", "comma-separated roles to activate")
	return cmd
}
