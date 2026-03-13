package tests

import (
	"testing"

	"github.com/momlesstomato/pixel-server/core/cli"
)

// TestNewRootCommandRegistersServe verifies root command composition.
func TestNewRootCommandRegistersServe(t *testing.T) {
	command := cli.NewRootCommand(cli.Dependencies{})
	serve, _, err := command.Find([]string{"serve"})
	if err != nil {
		t.Fatalf("expected serve command lookup success, got %v", err)
	}
	if serve == nil || serve.Name() != "serve" {
		t.Fatalf("expected serve command to be registered")
	}
}

// TestNewRootCommandRegistersSSO verifies sso command composition.
func TestNewRootCommandRegistersSSO(t *testing.T) {
	command := cli.NewRootCommand(cli.Dependencies{})
	sso, _, err := command.Find([]string{"sso"})
	if err != nil {
		t.Fatalf("expected sso command lookup success, got %v", err)
	}
	if sso == nil || sso.Name() != "sso" {
		t.Fatalf("expected sso command to be registered")
	}
}

// TestNewRootCommandRegistersDB verifies db command composition.
func TestNewRootCommandRegistersDB(t *testing.T) {
	command := cli.NewRootCommand(cli.Dependencies{})
	database, _, err := command.Find([]string{"db"})
	if err != nil {
		t.Fatalf("expected db command lookup success, got %v", err)
	}
	if database == nil || database.Name() != "db" {
		t.Fatalf("expected db command to be registered")
	}
}

// TestNewRootCommandRegistersUser verifies user command composition.
func TestNewRootCommandRegistersUser(t *testing.T) {
	command := cli.NewRootCommand(cli.Dependencies{})
	user, _, err := command.Find([]string{"user"})
	if err != nil {
		t.Fatalf("expected user command lookup success, got %v", err)
	}
	if user == nil || user.Name() != "user" {
		t.Fatalf("expected user command to be registered")
	}
}
