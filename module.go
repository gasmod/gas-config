package config

import (
	"sync/atomic"

	"github.com/gasmod/gas-config/providers"

	"github.com/go-playground/validator/v10"
)

// Option is a functional option for configuring a Module.
type Option func(*Config)

// WithProvider adds a configuration provider to the module.
func WithProvider(provider providers.Provider) Option {
	return func(m *Config) {
		m.providers = append(m.providers, provider)
	}
}

// WithExtension adds an extension with pre/post-load hooks to the module.
func WithExtension(extension Extension) Option {
	return func(m *Config) {
		m.extensions = append(m.extensions, extension)
	}
}

// New creates a new config module instance with given options.
func New(opts ...Option) *Config {
	m := &Config{
		values:     make(map[string]any),
		providers:  make([]providers.Provider, 0),
		extensions: make([]Extension, 0),
		validate:   validator.New(),
		closed:     atomic.Bool{},
	}

	for _, opt := range opts {
		opt(m)
	}

	hasEnvProvider := false

	for _, p := range m.providers {
		if p.Name() == providers.EnvProviderName {
			hasEnvProvider = true
			break
		}
	}

	if !hasEnvProvider {
		m.providers = append([]providers.Provider{providers.NewEnvProvider()}, m.providers...)
	}

	return m
}

// Name returns the module identifier.
func (m *Config) Name() string {
	return "gas-config"
}

// Init loads configuration from all registered providers.
func (m *Config) Init() error {
	return m.load()
}

// Close gracefully shuts down the config module.
func (m *Config) Close() error {
	return nil
}
