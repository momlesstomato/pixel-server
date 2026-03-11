package redis

import "testing"

// TestNewClientBuildsConfiguredInstance verifies Redis client creation.
func TestNewClientBuildsConfiguredInstance(t *testing.T) {
	client, err := NewClient(Config{
		Address: "localhost:6379", DB: 2, PoolSize: 33,
	})
	if err != nil {
		t.Fatalf("expected client creation success, got %v", err)
	}
	options := client.Options()
	if options.Addr != "localhost:6379" || options.DB != 2 || options.PoolSize != 33 {
		t.Fatalf("unexpected redis options: %+v", options)
	}
	if closeErr := client.Close(); closeErr != nil {
		t.Fatalf("expected close success, got %v", closeErr)
	}
}

// TestNewClientRejectsMissingAddress verifies precondition validation.
func TestNewClientRejectsMissingAddress(t *testing.T) {
	if _, err := NewClient(Config{}); err == nil {
		t.Fatalf("expected client creation failure for missing address")
	}
}

// TestInitializerBuildsRedisClient verifies package-owned initializer behavior.
func TestInitializerBuildsRedisClient(t *testing.T) {
	client, err := (Initializer{}).InitializeRedis(Config{Address: "localhost:6379"})
	if err != nil {
		t.Fatalf("expected initializer success, got %v", err)
	}
	if closeErr := client.Close(); closeErr != nil {
		t.Fatalf("expected close success, got %v", closeErr)
	}
}

// TestInitializerRejectsEmptyConfig verifies config precondition checks.
func TestInitializerRejectsEmptyConfig(t *testing.T) {
	if _, err := (Initializer{}).InitializeRedis(Config{}); err == nil {
		t.Fatalf("expected initializer failure for missing address")
	}
}
