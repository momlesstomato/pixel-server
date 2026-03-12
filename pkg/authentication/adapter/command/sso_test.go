package command

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	miniredis "github.com/alicebob/miniredis/v2"
)

// TestExecuteSSOIssuesTicket verifies command issue flow and output payload.
func TestExecuteSSOIssuesTicket(t *testing.T) {
	server := startMiniRedis(t)
	output := bytes.NewBuffer(nil)
	envFile := writeEnvFile(t, server.Addr())
	err := ExecuteSSO(Options{EnvFile: envFile, UserID: 42}, output)
	if err != nil {
		t.Fatalf("expected command success, got %v", err)
	}
	var payload map[string]string
	if err := json.Unmarshal(bytes.TrimSpace(output.Bytes()), &payload); err != nil {
		t.Fatalf("expected output json decode success, got %v", err)
	}
	if payload["ticket"] == "" || payload["expires_at"] == "" {
		t.Fatalf("unexpected output payload: %+v", payload)
	}
	if value, getErr := server.Get("sso:" + payload["ticket"]); getErr != nil || value != "42" {
		t.Fatalf("expected redis ticket persistence, got value=%q err=%v", value, getErr)
	}
}

// TestNewSSOCommandRequiresUserID verifies required flag validation behavior.
func TestNewSSOCommandRequiresUserID(t *testing.T) {
	command := NewSSOCommand(Dependencies{Output: bytes.NewBuffer(nil)})
	command.SetArgs([]string{"--env-file", writeEnvFile(t, "127.0.0.1:6379")})
	if err := command.Execute(); err == nil {
		t.Fatalf("expected command failure for missing user-id")
	}
}

// writeEnvFile writes a valid command environment file.
func writeEnvFile(t *testing.T, redisAddress string) string {
	t.Helper()
	filePath := filepath.Join(t.TempDir(), ".env")
	content := []byte("APP_API_KEY=secret\nREDIS_ADDRESS=" + redisAddress + "\nPOSTGRES_DSN=postgres://postgres:postgres@127.0.0.1:5432/pixel_server?sslmode=disable\nUSERS_JWT_SECRET=secret\n")
	if err := os.WriteFile(filePath, content, 0o600); err != nil {
		t.Fatalf("expected env file write success, got %v", err)
	}
	return filePath
}

// startMiniRedis creates an isolated miniredis instance for tests.
func startMiniRedis(t *testing.T) *miniredis.Miniredis {
	t.Helper()
	server, err := miniredis.Run()
	if err != nil {
		t.Fatalf("expected miniredis startup, got %v", err)
	}
	t.Cleanup(server.Close)
	return server
}
