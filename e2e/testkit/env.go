package testkit

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// WriteServeEnvFile writes a valid env file for serve end-to-end tests.
func WriteServeEnvFile(t *testing.T, redisAddress string) string {
	t.Helper()
	address := redisAddress
	if address == "" {
		address = "localhost:6379"
	}
	filePath := filepath.Join(t.TempDir(), ".env")
	content := []byte(fmt.Sprintf("APP_BIND_IP=127.0.0.1\nAPP_PORT=3987\nAPP_API_KEY=test-key\nREDIS_ADDRESS=%s\nPOSTGRES_DSN=postgres://postgres:postgres@127.0.0.1:5432/pixel_server?sslmode=disable\nPOSTGRES_MIGRATION_AUTO_UP=false\nPOSTGRES_SEED_AUTO_UP=false\nUSERS_JWT_SECRET=secret\nLOGGING_LEVEL=debug\n", address))
	if err := os.WriteFile(filePath, content, 0o600); err != nil {
		t.Fatalf("write env file: %v", err)
	}
	return filePath
}
