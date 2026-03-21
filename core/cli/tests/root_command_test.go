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

// TestNewRootCommandRegistersGroup verifies group command composition.
func TestNewRootCommandRegistersGroup(t *testing.T) {
	command := cli.NewRootCommand(cli.Dependencies{})
	group, _, err := command.Find([]string{"group"})
	if err != nil {
		t.Fatalf("expected group command lookup success, got %v", err)
	}
	if group == nil || group.Name() != "group" {
		t.Fatalf("expected group command to be registered")
	}
}

// TestNewRootCommandRegistersFurniture verifies furniture command composition.
func TestNewRootCommandRegistersFurniture(t *testing.T) {
	command := cli.NewRootCommand(cli.Dependencies{})
	cmd, _, err := command.Find([]string{"furniture"})
	if err != nil || cmd == nil || cmd.Name() != "furniture" {
		t.Fatalf("expected furniture command to be registered, err=%v", err)
	}
}

// TestNewRootCommandRegistersInventory verifies inventory command composition.
func TestNewRootCommandRegistersInventory(t *testing.T) {
	command := cli.NewRootCommand(cli.Dependencies{})
	cmd, _, err := command.Find([]string{"inventory"})
	if err != nil || cmd == nil || cmd.Name() != "inventory" {
		t.Fatalf("expected inventory command to be registered, err=%v", err)
	}
}

// TestNewRootCommandRegistersCatalog verifies catalog command composition.
func TestNewRootCommandRegistersCatalog(t *testing.T) {
	command := cli.NewRootCommand(cli.Dependencies{})
	cmd, _, err := command.Find([]string{"catalog"})
	if err != nil || cmd == nil || cmd.Name() != "catalog" {
		t.Fatalf("expected catalog command to be registered, err=%v", err)
	}
}

// TestNewRootCommandRegistersEconomy verifies economy command composition.
func TestNewRootCommandRegistersEconomy(t *testing.T) {
	command := cli.NewRootCommand(cli.Dependencies{})
	cmd, _, err := command.Find([]string{"economy"})
	if err != nil || cmd == nil || cmd.Name() != "economy" {
		t.Fatalf("expected economy command to be registered, err=%v", err)
	}
}

// TestNewRootCommandRegistersSubscription verifies subscription command composition.
func TestNewRootCommandRegistersSubscription(t *testing.T) {
	command := cli.NewRootCommand(cli.Dependencies{})
	cmd, _, err := command.Find([]string{"subscription"})
	if err != nil || cmd == nil || cmd.Name() != "subscription" {
		t.Fatalf("expected subscription command to be registered, err=%v", err)
	}
}
