package config

import (
	"fmt"
	"time"
)

type JWTConfig struct {
	Secret string
	Expiry time.Duration
}

func loadJWTConfig() (JWTConfig, error) {
	secret, err := getRequiredEnv("JWT_SECRET")
	if err != nil {
		return JWTConfig{}, fmt.Errorf("failed to load JWT_SECRET: %v", err)
	}

	expiryStr := getEnv("JWT_EXPIRY", "1h")
	expiry, err := time.ParseDuration(expiryStr)
	if err != nil {
		return JWTConfig{}, fmt.Errorf("invalid JWT_EXPIRY %q: %w", expiryStr, err)
	}

	return JWTConfig{Secret: secret, Expiry: expiry}, nil
}
