package config

import "fmt"

// DatabaseConfig são as credenciais compartilhadas por todas as sources
// (ver ADR-006): em desenvolvimento, todos os bounded contexts vivem na
// mesma instância de Postgres, apenas em bancos de dados diferentes — só o
// nome do banco varia por source (ver SourceConfig).
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
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

	sslMode := getEnv("DB_SSLMODE", "disable")

	return DatabaseConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		SSLMode:  sslMode,
	}, nil
}

func (d DatabaseConfig) DSN(dbName string) string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, dbName, d.SSLMode)
}
