package config

import "sync"

type Config struct {
	Env       string
	HTTP      HTTPConfig
	Database  DatabaseConfig
	Messaging MessagingConfig
}

var (
	instance *Config
	once     sync.Once
)

func Get() *Config {
	once.Do(func() {
		instance = &Config{}

		instance.Env = "development"

		instance.HTTP = LoadHTTPConfig()
		instance.Database = LoadDatabaseConfig()
		instance.Messaging = LoadMessagingConfig()
	})

	return instance
}
