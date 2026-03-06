package cli

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"pixelsv/pkg/config"
	httpserver "pixelsv/pkg/http"
	logpkg "pixelsv/pkg/log"
)

func newServeCommand() *cobra.Command {
	var envFile string
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "start core HTTP/WebSocket runtime",
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := config.NewViper(config.LoadOptions{EnvFile: envFile})
			if err != nil {
				return err
			}
			if err := logpkg.BindViper(v); err != nil {
				return err
			}
			if err := httpserver.BindViper(v); err != nil {
				return err
			}
			if _, err := config.FromViper(v); err != nil {
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
			httpCfg, err := httpserver.FromViper(v)
			if err != nil {
				return err
			}
			srv, err := httpserver.New(httpCfg, logger)
			if err != nil {
				return err
			}
			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			return srv.ListenAndServe(ctx)
		},
	}
	cmd.Flags().StringVar(&envFile, "env-file", ".env", "env file path")
	return cmd
}
