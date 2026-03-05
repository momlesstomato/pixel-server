package redis

import "fmt"

// NamespacedKey returns a namespaced key.
func NamespacedKey(prefix string, key string) string {
	return fmt.Sprintf("%s:%s", prefix, key)
}
