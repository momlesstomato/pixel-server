package postgres

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewClient creates a GORM PostgreSQL client from application configuration.
func NewClient(postgreSQLConfig Config) (*gorm.DB, error) {
	dsn := strings.TrimSpace(postgreSQLConfig.DSN)
	if dsn == "" {
		return nil, fmt.Errorf("postgres dsn is required")
	}
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{DisableAutomaticPing: true})
	if err != nil {
		return nil, err
	}
	sqlDatabase, err := database.DB()
	if err != nil {
		return nil, err
	}
	sqlDatabase.SetMaxOpenConns(postgreSQLConfig.MaxOpenConns)
	sqlDatabase.SetMaxIdleConns(postgreSQLConfig.MaxIdleConns)
	sqlDatabase.SetConnMaxLifetime(time.Duration(postgreSQLConfig.ConnMaxLifetimeSeconds) * time.Second)
	return database, nil
}
