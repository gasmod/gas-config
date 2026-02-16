// Example usage of the config package
package main

import (
	"fmt"
	"testing/fstest"

	config "github.com/gasmod/gas-config"
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
	fsys := fstest.MapFS{
		"config.json": &fstest.MapFile{
			Data: []byte(`{
			  "database": {
				"host": "localhost",
				"port": 5432,
				"user": "admin",
				"password": "admin"
			  },
			  "server": {
				"host": "0.0.0.0",
				"port": 8080
			  },
			  "logging": {
				"level": "debug"
			  }
			}`),
		},
	}

	// initialize config instance
	cfg := config.New(
		config.WithProvider(
			config.NewJSONProvider(
				config.WithJSONFilePath("config.json"),
				config.WithJSONFileFS(&fsys),
			),
		),
	)
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
