---
name: gas-config
description: >
  Reference documentation for the gas-config Go package
  (github.com/gasmod/gas-config) — configuration management service for the
  Gas ecosystem. Use this skill when writing, reviewing, or debugging Go code
  involving configuration loading, providers (env, JSON, .env), struct binding,
  hierarchical dot-notation access, ConfigProvider interface, extensions with
  pre/post-load hooks, gas-env environment detection, SetDefault/SetDefaults,
  or DI wiring of config into Gas services.
---

# Gas Config Package Reference

Configuration management service for the Gas ecosystem. Supports loading from
multiple providers (environment variables, JSON files, `.env` files) and
binding to Go structs.

```go
import config "github.com/gasmod/gas-config"
import "github.com/gasmod/gas-config/providers"
```

## Constructor

```go
// Option configures the config service constructor.
type Option func(*Config)

func New(opts ...Option) func() *Config

func WithProvider(p providers.Provider) Option
func WithExtension(ext Extension) Option
```

`New` returns a curried DI-injectable constructor (`func() *Config`).
An `EnvProvider` is automatically prepended if no provider named
`"Environment Variables"` is present.

## Config (Service)

Implements `gas.Service` and `gas.ConfigProvider`.

### Lifecycle

```go
func (c *Config) Name() string   // "gas-config"
func (c *Config) Init() error    // loads config from all registered providers
func (c *Config) Close() error   // no-op, graceful shutdown placeholder
```

### Reading values

```go
func (c *Config) Get(key string) any                       // dot notation: "database.host"
func (c *Config) Find(key string) (value any, exist bool)  // same, with existence check
func (c *Config) Values() map[string]any                   // deep-cloned snapshot
```

`Get` and `Find` support hierarchical dot-notation paths. Keys are
lowercased internally. Returns deep-cloned values (safe for mutation).

### Setting values

```go
func (c *Config) Set(key string, value any)           // overrides existing
func (c *Config) SetDefault(key string, value any)    // won't override existing
func (c *Config) SetDefaults(values any) error        // from struct/map, won't override
```

`Set` and `SetDefault` create nested maps along the dot-path if they don't
exist. Empty keys are ignored. `SetDefaults` accepts `*struct`, `*map[string]any`,
or `map[string]any`.

### Struct binding

```go
func (c *Config) Bind(dest any, options ...BindOption) error

type BindOption func(*BindOptions)

func WithValidate(validate bool) BindOption
```

Field matching: `json` tags first, then case-insensitive field names.
Validation uses `go-playground/validator` tags (enabled by default).

Supported types: all int/uint/float variants, bool, string, `time.Duration`,
slices (including comma-separated strings), maps, nested structs, embedded
structs.

## Providers

### Provider interface

```go
// In package providers
type Provider interface {
    Name() string
    Load() (map[string]any, error)
}
```

### Provider ordering

Later providers override earlier ones. The auto-registered `EnvProvider` is
prepended (lowest priority):

```
EnvProvider (auto, lowest) → user providers in order (highest last)
```

### Environment variables (default)

```go
providers.NewEnvProvider(opts ...EnvOption) *EnvProvider
```

Options:

```go
providers.WithEnvPrefix(prefix string)  // filter by prefix, stripped from keys
providers.WithEnvSeparator(sep string)  // nesting separator (default "__")
```

Env vars are lowercased and kept as flat `snake_case` keys by default:
`DATABASE_HOST` → `database_host`

Use `__` (double underscore) for nested map creation:
`DATABASE__HOST` → `database.host`

Constant: `providers.EnvProviderName = "Environment Variables"`

### JSON files

```go
providers.NewJSONProvider(opts ...JSONOption) *JSONProvider
```

Options:

```go
providers.WithJSONFilePath(filePath string)  // required
providers.WithJSONFileFS(fileFS fs.FS)       // optional: custom fs.FS (e.g. embed.FS)
```

Errors: `providers.ErrJSONFilePathNotSet`, `providers.ErrJSONFileReadFailed`,
`providers.ErrJSONDecodeFailed`

### .env files

```go
providers.NewDotEnvProvider(opts ...DotEnvOption) *DotEnvProvider
```

Options:

```go
providers.WithDotEnvFilePath(filePath string)               // default: ".env"
providers.WithDotEnvSeparator(sep string)                   // nesting separator (default "__")
providers.WithDotEnvFileFS(fileFS fs.FS)                    // custom filesystem
providers.WithDotEnvFileNotFoundPanic(panicIfNotFound bool)  // default: true
providers.WithDotEnvFileAppendToOSEnv(appendToOSEnv bool)    // inject into os.Environ
```

Same key conventions as `EnvProvider`: flat `snake_case` by default, `__` for nesting.

Errors: `providers.ErrDotEnvFilePathNotSet`, `providers.ErrDotEnvFileReadFailed`,
`providers.ErrDotEnvParseFailed`, `providers.ErrSetEnv`

## Extensions

```go
type Extension interface {
    Name() string
    PreLoad(ctx context.Context, cfg *Config) error
    PostLoad(ctx context.Context, cfg *Config) error
}
```

Extensions run hooks around provider loading:
1. All `PreLoad` hooks execute (in registration order)
2. All providers load (in registration order, later overrides earlier)
3. All `PostLoad` hooks execute (in registration order)

### gas-env Extension

```go
import gasenv "github.com/gasmod/gas-config/extensions/gas-env"
```

Environment detection and validation. Resolves current environment from
config providers, OS variable, or default.

#### Constructor

```go
gasenv.NewExtension(opts ...EnvOption) *Extension
```

Options:

```go
gasenv.WithEnvVarName(name string)          // OS var to check (default: "GAS_ENV")
gasenv.WithDefault(env Environment)         // fallback env (default: Development)
gasenv.WithConfigKey(key string)            // config map key (default: "GasEnv")
gasenv.WithAllowedEnvs(envs ...Environment) // valid envs (default: all four)
```

#### Environment type and constants

```go
type Environment string

const (
    Development Environment = "development"
    Testing     Environment = "testing"
    Staging     Environment = "staging"
    Production  Environment = "production"
)
```

Methods on `Environment`: `IsDevelopment()`, `IsTesting()`, `IsStaging()`,
`IsProduction()`, `IsDevelopmentLike()`, `IsProductionLike()`, `String()`.

#### Extension API

```go
func (em *Extension) Name() string             // "GasEnv"
func (em *Extension) Current() Environment
func (em *Extension) Is(env Environment) bool
func (em *Extension) IsDevelopment() bool
func (em *Extension) IsTesting() bool
func (em *Extension) IsStaging() bool
func (em *Extension) IsProduction() bool
func (em *Extension) IsDevelopmentLike() bool   // development || testing
func (em *Extension) IsProductionLike() bool    // production || staging
```

#### Resolution priority

1. Config providers (JSON, `.env`, etc.) — checked in `PostLoad`
2. OS environment variable (`GAS_ENV` by default) — checked in `PreLoad`
3. Default environment (`Development`)

Invalid values at any stage fall back to the default.

#### Constants

```go
const ExtensionName    = "GasEnv"
const DefaultConfigKey = "GasEnv"
const DefaultEnvVarName = "GAS_ENV"
const DefaultEnvironment = Development
```

#### Embedding in config structs

```go
type WithGasEnv struct {
    GasEnv Environment
}
```

Embed `gasenv.WithGasEnv` in your config struct. The `GasEnv` field is
auto-populated by `Bind()` from the config key.

```go
type AppConfig struct {
    gasenv.WithGasEnv
    Database struct {
        Host string
        Port int
    }
}

var appCfg AppConfig
cfg.Bind(&appCfg)
appCfg.GasEnv.IsProduction() // false
```

## Errors

```go
var (
    ErrProviderLoadFailed          = errors.New("failed to load from provider")
    ErrExtensionPreLoadHookFailed  = errors.New("failed to execute extension pre-load hook")
    ErrExtensionPostLoadHookFailed = errors.New("failed to execute extension post-load hook")
    ErrNilValues                   = errors.New("values cannot be nil")
)
```

## DI Wiring

### With Gas App

```go
app := gas.NewApp(
    gas.WithService[*config.Config](config.New(
        config.WithProvider(providers.NewJSONProvider(
            providers.WithJSONFilePath("config.json"),
        )),
        config.WithProvider(providers.NewDotEnvProvider()),
    ), gas.ServiceLifetimeSingleton),
)
```

### Standalone (no Gas App)

```go
cfg := config.New(
    config.WithProvider(providers.NewJSONProvider(
        providers.WithJSONFilePath("config.json"),
    )),
)()

if err := cfg.Init(); err != nil {
    log.Fatal(err)
}
defer cfg.Close()

var appCfg AppConfig
if err := cfg.Bind(&appCfg); err != nil {
    log.Fatal(err)
}
```

### With gas-env extension

```go
envExt := gasenv.NewExtension()

cfg := config.New(
    config.WithProvider(providers.NewDotEnvProvider()),
    config.WithExtension(envExt),
)()

if err := cfg.Init(); err != nil {
    log.Fatal(err)
}

envExt.Current()           // "development"
envExt.IsProduction()      // false
envExt.IsDevelopmentLike() // true
```

### Consuming via gas.ConfigProvider

Other services receive config through the provider interface without
importing gas-config:

```go
type Service struct {
    cfg gas.ConfigProvider
}

func New(cfg gas.ConfigProvider) *Service {
    return &Service{cfg: cfg}
}

func (s *Service) Init() error {
    var myCfg MyConfig
    return s.cfg.Bind(&myCfg)
}
```

### Auto-binding in other Gas packages

When packages like `gas-database` don't receive explicit `WithConfig`, they
automatically bind from the injected `gas.ConfigProvider`. This lets you
drive all service configuration from a single `.env` or JSON file without
manual wiring.
