package command

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/momlesstomato/pixel-server/core/config"
	postgrescore "github.com/momlesstomato/pixel-server/core/postgres"
	userapplication "github.com/momlesstomato/pixel-server/pkg/user/application"
	"github.com/momlesstomato/pixel-server/pkg/user/domain"
	userstore "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/store"
	"github.com/spf13/cobra"
)

// newGetCommand creates the user get subcommand.
func newGetCommand(deps Dependencies, options *options) *cobra.Command {
	return &cobra.Command{Use: "get [id]", Short: "Get user profile by ID", Args: cobra.ExactArgs(1), RunE: func(_ *cobra.Command, args []string) error {
		service, cleanup, err := openService(*options)
		if err != nil {
			return err
		}
		defer cleanup()
		userID, err := parsePositiveID(args[0])
		if err != nil {
			return err
		}
		value, err := service.FindByID(context.Background(), userID)
		if err != nil {
			return err
		}
		return printJSON(deps.Output, value)
	}}
}

// newUpdateCommand creates the user update subcommand.
func newUpdateCommand(deps Dependencies, options *options) *cobra.Command {
	var motto, figure, gender string
	var homeRoomID int
	var command *cobra.Command
	command = &cobra.Command{Use: "update [id]", Short: "Update user profile fields", Args: cobra.ExactArgs(1), RunE: func(_ *cobra.Command, args []string) error {
		service, cleanup, err := openService(*options)
		if err != nil {
			return err
		}
		defer cleanup()
		userID, err := parsePositiveID(args[0])
		if err != nil {
			return err
		}
		patch := domain.ProfilePatch{}
		if command.Flags().Changed("motto") {
			patch.Motto = &motto
		}
		if command.Flags().Changed("figure") {
			patch.Figure = &figure
		}
		if command.Flags().Changed("gender") {
			patch.Gender = &gender
		}
		if command.Flags().Changed("home-room-id") {
			patch.HomeRoomID = &homeRoomID
		}
		value, err := service.UpdateProfile(context.Background(), userID, patch)
		if err != nil {
			return err
		}
		return printJSON(deps.Output, value)
	}}
	command.Flags().StringVar(&motto, "motto", "", "New user motto")
	command.Flags().StringVar(&figure, "figure", "", "New user figure")
	command.Flags().StringVar(&gender, "gender", "", "New user gender")
	command.Flags().IntVar(&homeRoomID, "home-room-id", -1, "New home room id")
	return command
}

// openService resolves user service dependencies.
func openService(options options) (*userapplication.Service, func(), error) {
	loaded, err := config.Load(config.LoaderOptions{EnvFile: options.EnvFile, EnvPrefix: options.EnvPrefix})
	if err != nil {
		return nil, nil, err
	}
	database, err := postgrescore.NewClient(loaded.PostgreSQL)
	if err != nil {
		return nil, nil, err
	}
	sqlDatabase, err := database.DB()
	if err != nil {
		return nil, nil, err
	}
	repository, err := userstore.NewRepository(database)
	if err != nil {
		_ = sqlDatabase.Close()
		return nil, nil, err
	}
	service, err := userapplication.NewService(repository)
	if err != nil {
		_ = sqlDatabase.Close()
		return nil, nil, err
	}
	return service, func() { _ = sqlDatabase.Close() }, nil
}

// parsePositiveID parses one positive integer identifier.
func parsePositiveID(value string) (int, error) {
	id, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("id must be a positive integer")
	}
	return id, nil
}

// printJSON prints one payload as json.
func printJSON(output io.Writer, value any) error {
	writer := output
	if writer == nil {
		writer = os.Stdout
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(writer, string(payload))
	return err
}
