package config

import (
	"fmt"
	"os"
)

func getEnv(key string, fallback string) string {
	value, ok := lookupNonEmptyEnv(key)
	if !ok {
		return fallback
	}
	return value
}

func getRequiredEnv(key string) (string, error) {
	value, ok := lookupNonEmptyEnv(key)
	if !ok {
		return "", fmt.Errorf("environment variable %s is required but not set or empty", key)
	}
	return value, nil
}

func lookupNonEmptyEnv(key string) (string, bool) {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		return "", false
	}
	return value, true
}
