package config

import (
	"log"
	"sync"
)

type Config struct {
	Env            string
	HTTP           *HTTPServerConfig
	Database       *DatabaseConfig
	JWT            JWTConfig
	BootstrapAdmin BootstrapAdminConfig
}

var (
	instance *Config
	once     sync.Once
)

func Get() *Config {
	once.Do(func() {
		cfg, err := loadConfig()
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}
		instance = cfg
	})
	return instance
}

func loadConfig() (*Config, error) {
	env := getEnv("APP_ENV", "development")

	httpConfig, err := loadHTTPConfig()
	if err != nil {
		return nil, err
	}

	dbConfig, err := loadDatabaseConfig()
	if err != nil {
		return nil, err
	}

	jwtConfig, err := loadJWTConfig()
	if err != nil {
		return nil, err
	}

	bootstrapAdmin, err := loadBootstrapAdminConfig()
	if err != nil {
		return nil, err
	}

	return &Config{
		Env:            env,
		HTTP:           httpConfig,
		Database:       &dbConfig,
		JWT:            jwtConfig,
		BootstrapAdmin: bootstrapAdmin,
	}, nil
}
