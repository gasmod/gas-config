# gas-config

Configuration management service for the [Gas ecosystem](https://github.com/gasmod). Supports loading from multiple providers (environment variables, JSON files, `.env` files) and binding to Go structs.

## Features

- **Gas Service interface**: Implements `Name()`, `Init()`, `Close()` for seamless integration with the Gas DI container
- **DI-compatible constructor**: `New()` returns a curried constructor for use with `gas.WithService`
- **Multiple providers**: Environment variables, JSON files, `.env` files, and custom providers
- **Hierarchical config**: Nested access via dot notation (e.g., `database.host`)
- **Struct binding**: Type-safe binding with validation support
- **Sensible defaults**: Environment provider is registered automatically

## Installation

```bash
go get github.com/gasmod/gas-config
```

## Usage with Gas

```go
package main

import (
    "github.com/gasmod/gas"
    config "github.com/gasmod/gas-config"
    "github.com/gasmod/gas-config/providers"
)

func main() {
    app := gas.NewApp(
        gas.WithService[*config.Config](config.New(
            config.WithProvider(
                providers.NewJSONProvider(
                    providers.WithJSONFilePath("config.json"),
                ),
            ),
            config.WithProvider(providers.NewDotEnvProvider()),
        ), gas.ServiceLifetimeSingleton),
    )

    app.Run()
}
```

## Standalone usage

```go
package main

import (
    config "github.com/gasmod/gas-config"
    "github.com/gasmod/gas-config/providers"
)

func main() {
    // Create the config service with providers
    cfg := config.New(
        config.WithProvider(
            providers.NewJSONProvider(
                providers.WithJSONFilePath("config.json"),
            ),
        ),
        config.WithProvider(providers.NewDotEnvProvider()),
    )()

    // Load configuration
    if err := cfg.Init(); err != nil {
        panic(err)
    }

    // Bind to a struct
    var appCfg AppConfig
    if err := cfg.Bind(&appCfg); err != nil {
        panic(err)
    }
}

type AppConfig struct {
    Database struct {
        Host string
        Port int
    }
    Server struct {
        Host string
        Port int
    }
}
```

## Providers

### Environment variables (default)

An `EnvProvider` is automatically added if none is provided. Environment variables map to nested config via `_` separators:

```bash
export DATABASE_HOST=localhost
export DATABASE_PORT=5432
```

```go
cfg := config.New()() // EnvProvider included by default
```

Options:

```go
config.New(
    config.WithProvider(
        providers.NewEnvProvider(
            providers.WithEnvPrefix("APP"),
            providers.WithEnvSeparator("__"),
            providers.WithEnvNormalizeVarNames(true),
        ),
    ),
)()
```

### JSON files

```go
config.New(
    config.WithProvider(
        providers.NewJSONProvider(
            providers.WithJSONFilePath("config.json"),
            providers.WithJSONFileFS(embeddedFS), // optional: custom fs.FS
        ),
    ),
)()
```

### `.env` files

```go
config.New(
    config.WithProvider(
        providers.NewDotEnvProvider(
            providers.WithDotEnvFilePath(".env"),
            providers.WithDotEnvFileNotFoundPanic(false),
            providers.WithDotEnvFileAppendToOSEnv(true),
        ),
    ),
)()
```

### Custom providers

Implement the `Provider` interface:

```go
type Provider interface {
    Name() string
    Load() (map[string]any, error)
}
```

## Provider ordering

Later providers override earlier ones. The auto-registered `EnvProvider` is prepended, so explicit providers take precedence:

```go
cfg := config.New(
    config.WithProvider(providers.NewJSONProvider(...)),    // base config
    config.WithProvider(providers.NewDotEnvProvider()),     // overrides JSON
    // EnvProvider is prepended automatically (lowest priority)
)()
```

## Reading values

```go
// Get a single value
host := cfg.Get("database.host")

// Check if a value exists
val, exists := cfg.Find("database.host")

// Set defaults (won't override loaded values)
cfg.SetDefault("database.port", 5432)

// Set a value (overrides loaded values)
cfg.Set("database.host", "127.0.0.1")

// Get all values
allValues := cfg.Values()
```

## Struct binding

`Bind()` maps configuration into structs using reflection. Field matching uses `json` tags first, then case-insensitive field names:

```go
type DBConfig struct {
    Host     string `json:"host"`
    Port     int    `json:"port"`
    Password string `json:"password" validate:"required"`
}

var dbCfg DBConfig
if err := cfg.Bind(&dbCfg); err != nil {
    log.Fatal(err)
}

// Disable validation
cfg.Bind(&dbCfg, config.WithValidate(false))
```

Supported types: all int/uint/float variants, bool, string, `time.Duration`, slices (including comma-separated strings), maps, nested structs, and embedded structs.

## Extensions

Extensions provide pre/post-load hooks:

```go
type Extension interface {
    Name() string
    PreLoad(ctx context.Context, cfg *Config) error
    PostLoad(ctx context.Context, cfg *Config) error
}
```

```go
cfg := config.New(config.WithExtension(myExtension))()
```

### gas-env

The `gas-env` extension (`extensions/gas-env`) manages application environment detection and validation. It resolves the current environment from config providers, the `GAS_ENV` OS variable, or a default, and makes it available throughout the app.

```go
import (
    config "github.com/gasmod/gas-config"
    gasenv "github.com/gasmod/gas-config/extensions/gas-env"
)

envExt := gasenv.NewExtension()

cfg := config.New(
    config.WithProvider(providers.NewDotEnvProvider()),
    config.WithExtension(envExt),
)()

if err := cfg.Init(); err != nil {
    panic(err)
}

fmt.Println(envExt.Current())       // "development"
fmt.Println(envExt.IsProduction())  // false
fmt.Println(envExt.IsDevelopmentLike()) // true
```

**Environments:** `Development`, `Testing`, `Staging`, `Production`

**Resolution priority:**
1. Config providers (JSON, `.env`, etc.)
2. OS environment variable (`GAS_ENV` by default)
3. Default (`Development`)

**Options:**

```go
gasenv.NewExtension(
    gasenv.WithEnvVarName("APP_ENV"),
    gasenv.WithDefault(gasenv.Production),
    gasenv.WithConfigKey("AppEnv"),
    gasenv.WithAllowedEnvs(gasenv.Production, gasenv.Staging),
)
```

**Embedding in config structs:**

```go
type AppConfig struct {
    gasenv.WithGasEnv // adds GasEnv field, auto-populated by Bind()
    Database struct {
        Host string
        Port int
    }
}

var appCfg AppConfig
cfg.Bind(&appCfg)
fmt.Println(appCfg.GasEnv.IsProduction()) // false
```

## Examples

See [`examples/`](./examples/) for complete examples:

- [`basic/`](./examples/basic/) - Environment variables
- [`env/`](./examples/env/) - Environment variables with prefix
- [`json/`](./examples/json/) - JSON configuration
- [`json_fs/`](./examples/json_fs/) - JSON with embedded filesystem
- [`dotenv/`](./examples/dotenv/) - `.env` files
