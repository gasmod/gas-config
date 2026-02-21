package config

import (
	"sync/atomic"

	"github.com/gasmod/gas-config/providers"

	"github.com/go-playground/validator/v10"
)

// serviceName specifies the identifier for this service.
const serviceName = "gas-config"

// New returns a DI constructor for the config service.
// Providers define configuration sources (env, JSON, .env files, etc.).
// Extensions provide pre/post-load hooks.
// An EnvProvider is automatically prepended if none is provided.
func New(provs []providers.Provider, extensions []Extension) func() *Config {
	return func() *Config {
		if provs == nil {
			provs = make([]providers.Provider, 0)
		}

		if extensions == nil {
			extensions = make([]Extension, 0)
		}

		c := &Config{
			values:     make(map[string]any),
			providers:  provs,
			extensions: extensions,
			validate:   validator.New(),
			closed:     atomic.Bool{},
		}

		hasEnvProvider := false

		for _, p := range c.providers {
			if p.Name() == providers.EnvProviderName {
				hasEnvProvider = true
				break
			}
		}

		if !hasEnvProvider {
			c.providers = append([]providers.Provider{providers.NewEnvProvider()}, c.providers...)
		}

		return c
	}
}

// Name returns the service identifier.
func (c *Config) Name() string {
	return serviceName
}

// Init loads configuration from all registered providers.
func (c *Config) Init() error {
	return c.load()
}

// Close gracefully shuts down the config service.
func (c *Config) Close() error {
	return nil
}
