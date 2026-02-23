package config

import (
	"sync/atomic"

	"github.com/gasmod/gas-config/providers"

	"github.com/go-playground/validator/v10"
)

// serviceName specifies the identifier for this service.
const serviceName = "gas-config"

// Option configures the config service constructor.
type Option func(*Config)

// WithProvider adds a configuration provider (env, JSON, .env, etc.).
func WithProvider(p providers.Provider) Option {
	return func(c *Config) {
		c.providers = append(c.providers, p)
	}
}

// WithExtension registers an extension that provides pre/post-load hooks.
func WithExtension(ext Extension) Option {
	return func(c *Config) {
		c.extensions = append(c.extensions, ext)
	}
}

// New returns a DI constructor for the config service.
// An EnvProvider is automatically prepended if none is provided.
func New(opts ...Option) func() *Config {
	return func() *Config {
		c := &Config{
			values:     make(map[string]any),
			providers:  make([]providers.Provider, 0),
			extensions: make([]Extension, 0),
			validate:   validator.New(),
			closed:     atomic.Bool{},
		}

		for _, opt := range opts {
			opt(c)
		}

		if len(c.providers) == 0 {
			c.providers = append(c.providers, providers.NewEnvProvider())
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
