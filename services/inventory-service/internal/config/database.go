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
		return DatabaseConfig{}, fmt.Errorf("failed to load DB_HOST: %w", err)
	}

	port, err := getRequiredEnv("DB_PORT")
	if err != nil {
		return DatabaseConfig{}, fmt.Errorf("failed to load DB_PORT: %w", err)
	}

	user, err := getRequiredEnv("DB_USER")
	if err != nil {
		return DatabaseConfig{}, fmt.Errorf("failed to load DB_USER: %w", err)
	}

	password, err := getRequiredEnv("DB_PASSWORD")
	if err != nil {
		return DatabaseConfig{}, fmt.Errorf("failed to load DB_PASSWORD: %w", err)
	}

	name, err := getRequiredEnv("DB_NAME")
	if err != nil {
		return DatabaseConfig{}, fmt.Errorf("failed to load DB_NAME: %w", err)
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
