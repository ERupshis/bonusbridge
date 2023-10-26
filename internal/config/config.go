// Package config server's setting parser. Applies flags and environments. Environments are prioritized.
package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env"
)

// Config server's settings.
type Config struct {
	AccrualAddr string // AccrualAddr accrual system address.
	DatabaseDSN string // DatabaseDSN PostgreSQL data source name.
	HostAddr    string // Host server's address.
	JWTKey      string // jwt web token generation key.
	LogLevel    string // log level.
}

// Parse main func to parse variables.
func Parse() Config {
	var config = Config{}
	checkFlags(&config)
	checkEnvironments(&config)

	config.AccrualAddr = AddHTTPPrefixIfNeed(config.AccrualAddr)
	return config
}

// FLAGS PARSING.
const (
	flagHostAddress    = "a"
	flagDatabaseDSN    = "d"
	flagAccrualAddress = "r"
	flagJWTKey         = "j"
	flagLogLevel       = "l"
)

// checkFlags checks flags of app's launch.
func checkFlags(config *Config) {
	// main app.
	flag.StringVar(&config.HostAddr, flagHostAddress, "localhost:8080", "server endpoint")

	// postgres.
	flag.StringVar(&config.DatabaseDSN, flagDatabaseDSN, "postgres://postgres:postgres@localhost:5432/gophermart_db?sslmode=disable", "database DSN")

	// accrual.
	flag.StringVar(&config.AccrualAddr, flagAccrualAddress, "localhost:8080", "accrual system address")

	// accrual.
	flag.StringVar(&config.JWTKey, flagJWTKey, "need TO REMOVE", "JWT web token key")

	// log.
	flag.StringVar(&config.LogLevel, flagLogLevel, "info", "log level")

	flag.Parse()
}

// ENVIRONMENTS PARSING.
// envConfig struct of environments suitable for server.
type envConfig struct {
	AccrualAddr string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	DatabaseDSN string `env:"DATABASE_URI"`
	HostAddr    string `env:"RUN_ADDRESS"`
	JWTKey      string `env:"JWT_KEY"`
	LogLevel    string `env:"LOG_LEVEL"`
}

// checkEnvironments checks environments suitable for server.
func checkEnvironments(config *Config) {
	var envs = envConfig{}
	err := env.Parse(&envs)
	if err != nil {
		log.Fatal(err)
	}

	// main app.
	_ = SetEnvToParamIfNeed(&config.HostAddr, envs.HostAddr)

	// postgres.
	_ = SetEnvToParamIfNeed(&config.DatabaseDSN, envs.DatabaseDSN)

	// accrual.
	_ = SetEnvToParamIfNeed(&config.AccrualAddr, envs.AccrualAddr)

	//authentication.
	_ = SetEnvToParamIfNeed(&config.JWTKey, envs.JWTKey)

	//log level.
	_ = SetEnvToParamIfNeed(&config.LogLevel, envs.LogLevel)
}
