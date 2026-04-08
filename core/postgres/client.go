package postgres

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// NewClient creates a GORM PostgreSQL client from application configuration.
func NewClient(postgreSQLConfig Config) (*gorm.DB, error) {
	dsn := strings.TrimSpace(postgreSQLConfig.DSN)
	if dsn == "" {
		return nil, fmt.Errorf("postgres dsn is required")
	}
	logger := gormlogger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), gormlogger.Config{
		LogLevel:                  gormlogger.Warn,
		IgnoreRecordNotFoundError: true,
		Colorful:                  false,
	})
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		DisableAutomaticPing: true,
		Logger:               logger,
	})
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
