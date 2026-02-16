package config

// Provider defines the interface for configuration providers.
// Implement this interface to create custom providers like env, json, yml, etc.
type Provider interface {
	Name() string
	// Load reads configuration from the source and returns it as a map.
	// Keys should be hierarchical paths (e.g., "database.host").
	Load() (map[string]any, error)
}
