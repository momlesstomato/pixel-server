package redis

import "testing"

// TestNamespacedKey checks key generation.
func TestNamespacedKey(t *testing.T) {
	key := NamespacedKey("px", "abc")
	if key != "px:abc" {
		t.Fatalf("unexpected key: %s", key)
	}
}
