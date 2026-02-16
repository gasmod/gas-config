# gas-config

Configuration management module for the [Gas ecosystem](https://github.com/gasmod). Supports loading from multiple providers (environment variables, JSON files, `.env` files) and binding to Go structs.

## Features

- **Gas Module interface**: Implements `Name()`, `Init()`, `Close()` for seamless integration with the Gas base server
- **Functional options**: Configure providers and extensions via `WithProvider()` and `WithExtension()`
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
    config "github.com/gasmod/gas-config"
)

func main() {
    // Create the config module with providers
    cfgMod := config.New(
        config.WithProvider(config.NewJSONProvider(
            config.WithJSONFilePath("config.json"),
        )),
        config.WithProvider(config.NewDotEnvProvider()),
    )

    // In a Gas app, the base server calls Init() automatically.
    // For standalone use:
    if err := cfgMod.Init(); err != nil {
        panic(err)
    }

    // Bind to a struct
    var cfg AppConfig
    if err := cfgMod.Bind(&cfg); err != nil {
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
cfgMod := config.New() // EnvProvider included by default
```

Options:

```go
config.NewEnvProvider(
    config.WithEnvPrefix("APP"),         // filter by prefix
    config.WithEnvSeparator("__"),       // custom separator for nesting
    config.WithEnvNormalizeVarNames(true), // snake_case to camelCase
)
```

### JSON files

```go
config.WithProvider(config.NewJSONProvider(
    config.WithJSONFilePath("config.json"),
    config.WithJSONFileFS(embeddedFS), // optional: custom fs.FS
))
```

### `.env` files

```go
config.WithProvider(config.NewDotEnvProvider(
    config.WithDotEnvFilePath(".env"),
    config.WithDotEnvFileNotFoundPanic(false), // silently skip if missing
    config.WithDotEnvFileAppendToOSEnv(true),  // add vars to os.Environ
))
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
cfgMod := config.New(
    config.WithProvider(config.NewJSONProvider(...)),  // base config
    config.WithProvider(config.NewDotEnvProvider()), // overrides JSON
    // EnvProvider is prepended automatically (lowest priority)
)
```

## Reading values

```go
// Get a single value
host := cfgMod.Get("database.host")

// Check if a value exists
val, exists := cfgMod.Find("database.host")

// Set defaults (won't override loaded values)
cfgMod.SetDefault("database.port", 5432)

// Set a value (overrides loaded values)
cfgMod.Set("database.host", "127.0.0.1")

// Get all values
allValues := cfgMod.Values()
```

## Struct binding

`Bind()` maps configuration into structs using reflection. Field matching uses `json` tags first, then case-insensitive field names:

```go
type DBConfig struct {
    Host     string `json:"host"`
    Port     int    `json:"port"`
    Password string `json:"password" validate:"required"`
}

var cfg DBConfig
if err := cfgMod.Bind(&cfg); err != nil {
    log.Fatal(err)
}

// Disable validation
cfgMod.Bind(&cfg, config.WithValidate(false))
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
cfgMod := config.New(
    config.WithExtension(myExtension),
)
```

### gas-env

The `gas-env` extension (`extensions/gas-env`) manages application environment detection and validation. It resolves the current environment from config providers, the `GAS_ENV` OS variable, or a default, and makes it available throughout the app.

```go
import (
    config "github.com/gasmod/gas-config"
    gasenv "github.com/gasmod/gas-config/extensions/gas-env"
)

envExt := gasenv.NewExtension()

cfgMod := config.New(
    config.WithExtension(envExt),
    config.WithProvider(config.NewDotEnvProvider()),
)

if err := cfgMod.Init(); err != nil {
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
    gasenv.WithEnvVarName("APP_ENV"),          // custom OS variable name
    gasenv.WithDefault(gasenv.Production),      // custom default
    gasenv.WithConfigKey("AppEnv"),             // custom config key (default: "GasEnv")
    gasenv.WithAllowedEnvs(gasenv.Production, gasenv.Staging), // restrict valid values
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

var cfg AppConfig
cfgMod.Bind(&cfg)
fmt.Println(cfg.GasEnv.IsProduction()) // false
```

## Examples

See [`examples/`](./examples/) for complete examples:

- [`basic/`](./examples/basic/) - Environment variables
- [`env/`](./examples/env/) - Environment variables with prefix
- [`json/`](./examples/json/) - JSON configuration
- [`json_fs/`](./examples/json_fs/) - JSON with embedded filesystem
- [`dotenv/`](./examples/dotenv/) - `.env` files
