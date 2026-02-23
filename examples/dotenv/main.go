// Example usage of the config package
package main

import (
	"fmt"

	config "github.com/gasmod/gas-config"
	"github.com/gasmod/gas-config/providers"
)

type AppConfig struct {
	Database struct {
		Host     string
		Port     int
		User     string
		Password string
	}
	Server struct {
		Host string
		Port int
	}
	Logging struct {
		Level string
	}
}

func main() {
	// initialize config service
	cfg := config.New(
		config.WithProvider(
			providers.NewDotEnvProvider(providers.WithDotEnvSeparator("__")),
		), // default DotEnv file path is ".env"
	)()
	defer cfg.Close()

	// Load configuration
	if err := cfg.Init(); err != nil {
		panic(err)
	}

	// Bind to user-defined type
	var appCfg AppConfig
	if err := cfg.Bind(&appCfg); err != nil {
		panic(err)
	}

	// Use the config
	fmt.Printf("Server: %s:%d\n", appCfg.Server.Host, appCfg.Server.Port)
	fmt.Printf(
		"DB: postgresql://%s:%s@%s:%d\n",
		appCfg.Database.User,
		appCfg.Database.Password,
		appCfg.Database.Host,
		appCfg.Database.Port,
	)
	fmt.Printf("Log Level: %s\n", appCfg.Logging.Level)
}
