package config

import "fmt"

type JWTConfig struct {
	Secret string
}

func loadJWTConfig() (JWTConfig, error) {
	secret, err := getRequiredEnv("JWT_SECRET")
	if err != nil {
		return JWTConfig{}, fmt.Errorf("failed to load JWT_SECRET: %v", err)
	}

	return JWTConfig{Secret: secret}, nil
}
