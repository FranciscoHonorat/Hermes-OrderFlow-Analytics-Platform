package config

import "fmt"

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

func loadDatabaseConfig() (DatabaseConfig, error) {
	host, err := getRequiredEnv("DB_HOST")
	if err != nil {
		return DatabaseConfig{}, fmt.Errorf("failed to load DB_HOST: %v", err)
	}

	port, err := getRequiredEnv("DB_PORT")
	if err != nil {
		return DatabaseConfig{}, fmt.Errorf("failed to load DB_PORT: %v", err)
	}

	user, err := getRequiredEnv("DB_USER")
	if err != nil {
		return DatabaseConfig{}, fmt.Errorf("failed to load DB_USER: %v", err)
	}

	password, err := getRequiredEnv("DB_PASSWORD")
	if err != nil {
		return DatabaseConfig{}, fmt.Errorf("failed to load DB_PASSWORD: %v", err)
	}

	name, err := getRequiredEnv("DB_NAME")
	if err != nil {
		return DatabaseConfig{}, fmt.Errorf("failed to load DB_NAME: %v", err)
	}

	sslMode := getEnv("DB_SSLMODE", "disable")

	return DatabaseConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Name:     name,
		SSLMode:  sslMode,
	}, nil
}
