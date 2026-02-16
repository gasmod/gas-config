package config

import (
	"sync/atomic"

	"github.com/go-playground/validator/v10"
)

// Module is the gas-config module that implements the Gas Module interface.
// It wraps Config with lifecycle management for the Gas ecosystem.
type Module struct {
	Config

	closed atomic.Bool
}

// Option is a functional option for configuring a Module.
type Option func(*Module)

// WithProvider adds a configuration provider to the module.
func WithProvider(provider Provider) Option {
	return func(m *Module) {
		m.providers = append(m.providers, provider)
	}
}

// WithExtension adds an extension with pre/post-load hooks to the module.
func WithExtension(extension Extension) Option {
	return func(m *Module) {
		m.extensions = append(m.extensions, extension)
	}
}

// New creates a new config module instance with given options.
func New(opts ...Option) *Module {
	m := &Module{
		Config: Config{
			values:     make(map[string]any),
			providers:  make([]Provider, 0),
			extensions: make([]Extension, 0),
			validate:   validator.New(),
		},
		closed: atomic.Bool{},
	}

	for _, opt := range opts {
		opt(m)
	}

	hasEnvProvider := false

	for _, p := range m.providers {
		if p.Name() == envProviderName {
			hasEnvProvider = true

			break
		}
	}

	if !hasEnvProvider {
		m.providers = append([]Provider{NewEnvProvider()}, m.providers...)
	}

	return m
}

// Name returns the module identifier.
func (m *Module) Name() string {
	return "gas-config"
}

// Init loads configuration from all registered providers.
func (m *Module) Init() error {
	return m.load()
}

// Close gracefully shuts down the config module.
func (m *Module) Close() error {
	return nil
}
