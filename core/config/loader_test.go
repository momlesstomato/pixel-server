package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLoadCombinesEnvFileAndEnvironment verifies .env loading and env override precedence.
func TestLoadCombinesEnvFileAndEnvironment(t *testing.T) {
	t.Setenv("CFG_TEST_APP_PORT", "9090")
	t.Setenv("CFG_TEST_USERS_JWT_SECRET", "from-env")
	envFile := writeEnvFile(t, strings.Join([]string{
		"APP_BIND_IP=127.0.0.1",
		"APP_PORT=8080",
		"REDIS_ADDRESS=localhost:6379",
		"POSTGRES_DSN=postgres://pixel:pixel@localhost:5432/pixel?sslmode=disable",
		"USERS_JWT_SECRET=from-file",
		"LOGGING_FORMAT=json",
	}, "\n"))
	loaded, err := Load(LoaderOptions{EnvFile: envFile, EnvPrefix: "CFG_TEST"})
	if err != nil {
		t.Fatalf("expected successful load, got error: %v", err)
	}
	if loaded.App.Port != 9090 {
		t.Fatalf("expected env override for app.port, got %d", loaded.App.Port)
	}
	if loaded.Users.JWTSecret != "from-env" {
		t.Fatalf("expected env override for users.jwt_secret, got %q", loaded.Users.JWTSecret)
	}
	if loaded.App.BindIP != "127.0.0.1" {
		t.Fatalf("expected file value for app.bind_ip, got %q", loaded.App.BindIP)
	}
	if loaded.Logging.Format != "json" {
		t.Fatalf("expected file value for logging.format, got %q", loaded.Logging.Format)
	}
	if loaded.Redis.DB != 0 {
		t.Fatalf("expected default redis.db, got %d", loaded.Redis.DB)
	}
}

// TestLoadFailsOnMissingMandatoryFields verifies startup failure for missing required fields.
func TestLoadFailsOnMissingMandatoryFields(t *testing.T) {
	envFile := writeEnvFile(t, "APP_PORT=8080")
	_, err := Load(LoaderOptions{EnvFile: envFile, EnvPrefix: "CFG_MISSING"})
	if err == nil {
		t.Fatalf("expected load to fail when mandatory fields are missing")
	}
	message := err.Error()
	required := []string{"CFG_MISSING_POSTGRES_DSN", "CFG_MISSING_REDIS_ADDRESS", "CFG_MISSING_USERS_JWT_SECRET"}
	for _, expected := range required {
		if !strings.Contains(message, expected) {
			t.Fatalf("expected error to contain %q, got: %s", expected, message)
		}
	}
}

// TestLoadWithoutEnvFile verifies loading from env vars without an existing file.
func TestLoadWithoutEnvFile(t *testing.T) {
	t.Setenv("CFG_ONLY_REDIS_ADDRESS", "localhost:6379")
	t.Setenv("CFG_ONLY_POSTGRES_DSN", "postgres://pixel:pixel@localhost:5432/pixel?sslmode=disable")
	t.Setenv("CFG_ONLY_USERS_JWT_SECRET", "secret")
	loaded, err := Load(LoaderOptions{
		EnvFile:   filepath.Join(t.TempDir(), "missing.env"),
		EnvPrefix: "CFG_ONLY",
	})
	if err != nil {
		t.Fatalf("expected successful load from env vars, got error: %v", err)
	}
	if loaded.App.Port != 3000 {
		t.Fatalf("expected default app.port, got %d", loaded.App.Port)
	}
	if loaded.Logging.Level != "info" {
		t.Fatalf("expected default logging.level, got %q", loaded.Logging.Level)
	}
}

// TestLoadFailsOnUnreadableConfigPath verifies non-file paths fail early.
func TestLoadFailsOnUnreadableConfigPath(t *testing.T) {
	_, err := Load(LoaderOptions{EnvFile: t.TempDir(), EnvPrefix: "CFG_BAD_PATH"})
	if err == nil {
		t.Fatalf("expected load to fail for unreadable config path")
	}
}

// writeEnvFile creates a temporary .env file with the provided content.
func writeEnvFile(t *testing.T, content string) string {
	t.Helper()
	filePath := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(filePath, []byte(content), 0o600); err != nil {
		t.Fatalf("write .env file: %v", err)
	}
	return filePath
}
