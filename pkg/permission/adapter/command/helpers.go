package command

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/momlesstomato/pixel-server/core/config"
	postgrescore "github.com/momlesstomato/pixel-server/core/postgres"
	permissionapplication "github.com/momlesstomato/pixel-server/pkg/permission/application"
	permissionstore "github.com/momlesstomato/pixel-server/pkg/permission/infrastructure/store"
)

// openService resolves permission service dependencies.
func openService(options options) (*permissionapplication.Service, func(), error) {
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
	repository, err := permissionstore.NewRepository(database)
	if err != nil {
		_ = sqlDatabase.Close()
		return nil, nil, err
	}
	service, err := permissionapplication.NewService(repository, nil, permissionapplication.Config{
		CachePrefix: loaded.Permission.CachePrefix, CacheTTL: time.Duration(loaded.Permission.CacheTTLSeconds) * time.Second,
		AmbassadorPermission: loaded.Permission.AmbassadorPermission,
	})
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

// runWithService runs one action with a resolved permission service.
func runWithService(options options, action func(context.Context, *permissionapplication.Service) error) error {
	service, cleanup, err := openService(options)
	if err != nil {
		return err
	}
	defer cleanup()
	return action(context.Background(), service)
}
