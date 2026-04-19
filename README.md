# gas-config

Configuration management service for the [Gas ecosystem](https://github.com/gasmod). Supports loading from multiple providers (environment variables, JSON files, `.env` files) and binding to Go structs.

## Features

- **Early initialization**: `New()` returns `*Config` directly — create and load before the app starts
- **Multiple providers**: Environment variables, JSON files, `.env` files, and custom providers
- **Hierarchical config**: Nested access via dot notation (e.g., `database.host`)
- **Struct binding**: Type-safe binding with validation support
- **Sensible defaults**: Environment provider is registered automatically

## Installation

```bash
go get github.com/gasmod/gas-config
```

## Usage

### Standalone

```go
package main

import (
    config "github.com/gasmod/gas-config"
    "github.com/gasmod/gas-config/providers"
)

func main() {
    // Create and load config before anything else
    cfg := config.New(
        config.WithProvider(
            providers.NewJSONProvider(
                providers.WithJSONFilePath("config.json"),
            ),
        ),
        config.WithProvider(providers.NewDotEnvProvider()),
    )

    if err := cfg.Load(); err != nil {
        panic(err)
    }

    // Bind to a struct
    var appCfg AppConfig
    if err := cfg.Bind(&appCfg); err != nil {
        panic(err)
    }
}
```

### With Gas App

Config is loaded before `app.Run()` and registered as a singleton instance so
other services can receive it as `gas.ConfigProvider` via constructor injection.

```go
package main

import (
    "github.com/gasmod/gas"
    config "github.com/gasmod/gas-config"
    "github.com/gasmod/gas-config/providers"
)

func main() {
    cfg := config.New(
        config.WithProvider(providers.NewDotEnvProvider()),
        config.WithProvider(
            providers.NewJSONProvider(
                providers.WithJSONFilePath("config.json"),
            ),
        ),
    )

    if err := cfg.Load(); err != nil {
        panic(err)
    }

    app := gas.NewApp(
        gas.WithServiceInstance[gas.ConfigProvider](cfg),
        // other services ...
    )

    app.Run()
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

An `EnvProvider` is automatically added if none is provided. Environment variables are lowercased and kept as flat `snake_case` keys by default:

```bash
export DATABASE_HOST=localhost   # → database_host
export DATABASE_PORT=5432        # → database_port
```

To create nested maps, use `__` (double underscore) as the separator:

```bash
export DATABASE__HOST=localhost  # → database.host (nested)
export DATABASE__PORT=5432       # → database.port (nested)
```

```go
cfg := config.New() // EnvProvider included by default
```

Options:

```go
config.New(
    config.WithProvider(
        providers.NewEnvProvider(
            providers.WithEnvPrefix("APP"),
            providers.WithEnvSeparator("__"),
        ),
    ),
)
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
)
```

### `.env` files

`.env` variables follow the same key conventions as env vars: lowercased, flat `snake_case` by default; use `__` for nesting.

```go
config.New(
    config.WithProvider(
        providers.NewDotEnvProvider(
            providers.WithDotEnvFilePath(".env"),
            providers.WithDotEnvFileNotFoundPanic(false),
            providers.WithDotEnvFileAppendToOSEnv(true),
        ),
    ),
)
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
)
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

### Custom validator

Pass `config.WithValidator` to `New` to use a caller-owned `*validator.Validate` (e.g. with custom tags registered). If omitted, a new `validator.New()` instance is used internally.

```go
v := validator.New()
v.RegisterValidation("mytag", myValidatorFunc)

cfg := config.New(config.WithValidator(v))
```

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
cfg := config.New(config.WithExtension(myExtension))
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
)

if err := cfg.Load(); err != nil {
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

## Testing

The `configtest` package provides a mock that structurally satisfies `gas.ConfigProvider`:

```go
import "github.com/gasmod/gas-config/configtest"

mock := &configtest.MockConfig{}
mock.GetFn = func(key string) any {
	if key == "database.host" {
		return "localhost"
	}
	return nil
}

// assert calls:
if mock.CallCount("Get") != 1 {
	t.Error("expected one Get call")
}
```

For tests that need real `Get`/`Find`/`Bind` semantics seeded with known values, use `NewMockConfigWithValues`, which delegates to a real `*config.Config` under the hood and avoids leaking real environment variables:

```go
mock, err := configtest.NewMockConfigWithValues(map[string]any{
	"database.host": "localhost",
	"database.port": 5432,
})
```

Individual `Fn` fields can still be overridden afterwards, and `Calls`/`Reset`/`CallCount` work as usual.

## Examples

See [`examples/`](./examples/) for complete examples:

- [`basic/`](./examples/basic/) - Environment variables
- [`env/`](./examples/env/) - Environment variables with prefix
- [`json/`](./examples/json/) - JSON configuration
- [`json_fs/`](./examples/json_fs/) - JSON with embedded filesystem
- [`dotenv/`](./examples/dotenv/) - `.env` files
