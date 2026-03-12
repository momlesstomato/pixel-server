package command

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/momlesstomato/pixel-server/core/config"
	"github.com/momlesstomato/pixel-server/core/redis"
	"github.com/momlesstomato/pixel-server/pkg/authentication/application"
	"github.com/momlesstomato/pixel-server/pkg/authentication/domain"
	"github.com/momlesstomato/pixel-server/pkg/authentication/infrastructure/redisstore"
	"github.com/spf13/cobra"
)

// Dependencies defines command runtime overrides.
type Dependencies struct {
	// Output defines command output destination.
	Output io.Writer
}

// Options defines command execution inputs.
type Options struct {
	// EnvFile defines configuration file path.
	EnvFile string
	// EnvPrefix defines optional configuration prefix.
	EnvPrefix string
	// UserID defines user identifier for ticket issuance.
	UserID int
	// TTL defines optional ticket lifetime override.
	TTL time.Duration
}

// NewSSOCommand creates the authentication SSO command.
func NewSSOCommand(dependencies Dependencies) *cobra.Command {
	var options Options
	command := &cobra.Command{
		Use:   "sso",
		Short: "Issue one authentication SSO ticket",
		RunE: func(_ *cobra.Command, _ []string) error {
			return ExecuteSSO(options, dependencies.Output)
		},
	}
	command.Flags().StringVar(&options.EnvFile, "env-file", ".env", "Environment file path")
	command.Flags().StringVar(&options.EnvPrefix, "env-prefix", "", "Environment key prefix")
	command.Flags().IntVar(&options.UserID, "user-id", 0, "User identifier")
	command.Flags().DurationVar(&options.TTL, "ttl", 0, "Optional ticket lifetime")
	_ = command.MarkFlagRequired("user-id")
	return command
}

// ExecuteSSO executes command logic and prints ticket payload as JSON.
func ExecuteSSO(options Options, output io.Writer) error {
	loaded, err := config.Load(config.LoaderOptions{EnvFile: options.EnvFile, EnvPrefix: options.EnvPrefix})
	if err != nil {
		return err
	}
	client, err := redis.NewClient(loaded.Redis)
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }()
	store, err := redisstore.NewRedisStore(client, loaded.Authentication.KeyPrefix)
	if err != nil {
		return err
	}
	service := application.NewService(store, loaded.Authentication)
	result, err := service.Issue(context.Background(), domain.IssueRequest{
		UserID: options.UserID, TTL: options.TTL,
	})
	if err != nil {
		return err
	}
	if output == nil {
		output = os.Stdout
	}
	payload, err := json.Marshal(struct {
		Ticket    string `json:"ticket"`
		ExpiresAt string `json:"expires_at"`
	}{Ticket: result.Ticket, ExpiresAt: result.ExpiresAt.UTC().Format(time.RFC3339)})
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(output, string(payload))
	return err
}
