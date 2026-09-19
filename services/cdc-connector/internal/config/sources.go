package config

import (
	"fmt"
	"strings"
)

// SourceConfig identifica um bounded context cuja tabela outbox o relay
// deve drenar: Name vira o prefixo do tópico Kafka (ADR-004), DBName é o
// banco de dados dentro da mesma instância Postgres (ADR-006).
type SourceConfig struct {
	Name   string
	DBName string
}

// loadSourcesConfig parseia SOURCES no formato "nome:banco,nome:banco",
// ex: "order-service:orderflow,inventory-service:inventory".
func loadSourcesConfig() ([]SourceConfig, error) {
	raw, err := getRequiredEnv("SOURCES")
	if err != nil {
		return nil, err
	}

	entries := strings.Split(raw, ",")
	sources := make([]SourceConfig, 0, len(entries))

	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		parts := strings.SplitN(entry, ":", 2)
		if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
			return nil, fmt.Errorf("invalid SOURCES entry %q: expected format \"name:database\"", entry)
		}

		sources = append(sources, SourceConfig{
			Name:   strings.TrimSpace(parts[0]),
			DBName: strings.TrimSpace(parts[1]),
		})
	}

	if len(sources) == 0 {
		return nil, fmt.Errorf("SOURCES must list at least one \"name:database\" entry")
	}

	return sources, nil
}
