package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnv(t *testing.T) {
	t.Run("Test all scenarios for getEnv", func(t *testing.T) {
		t.Run("Test default env", func(t *testing.T) {
			env := getEnv("APP_ENV", "development")
			assert.Equal(t, "development", env)
		})

		t.Run("Test set env", func(t *testing.T) {
			t.Setenv("APP_ENV", "production")
			env := getEnv("APP_ENV", "development")
			assert.Equal(t, "production", env)
		})

		t.Run("Test missing env", func(t *testing.T) {
			value, exists := os.LookupEnv("APP_ENV")

			t.Cleanup(func() {
				if exists {
					os.Setenv("APP_ENV", value)
				} else {
					os.Unsetenv("APP_ENV")
				}
			})

			os.Unsetenv("APP_ENV")
			env := getEnv("APP_ENV", "development")
			assert.Equal(t, "development", env)
		})

		t.Run("Test empty env", func(t *testing.T) {
			t.Setenv("APP_ENV", "")
			env := getEnv("APP_ENV", "development")
			assert.Equal(t, "development", env)
		})
	})

	t.Run("Test all scenarios for getRequiredEnv", func(t *testing.T) {
		t.Run("Test set env", func(t *testing.T) {
			t.Setenv("APP_ENV", "production")
			env, err := getRequiredEnv("APP_ENV")
			assert.NoError(t, err)
			assert.Equal(t, "production", env)
		})

		t.Run("Test missing env", func(t *testing.T) {
			value, exists := os.LookupEnv("APP_ENV")

			t.Cleanup(func() {
				if exists {
					os.Setenv("APP_ENV", value)
				} else {
					os.Unsetenv("APP_ENV")
				}
			})

			os.Unsetenv("APP_ENV")
			_, err := getRequiredEnv("APP_ENV")
			assert.Error(t, err)
		})

		t.Run("Test empty env", func(t *testing.T) {
			t.Setenv("APP_ENV", "")
			_, err := getRequiredEnv("APP_ENV")
			assert.Error(t, err)
		})
	})

	t.Run("Test loadJWTConfig requires JWT_SECRET", func(t *testing.T) {
		t.Run("Missing secret fails closed", func(t *testing.T) {
			value, exists := os.LookupEnv("JWT_SECRET")
			t.Cleanup(func() {
				if exists {
					os.Setenv("JWT_SECRET", value)
				} else {
					os.Unsetenv("JWT_SECRET")
				}
			})

			os.Unsetenv("JWT_SECRET")
			_, err := loadJWTConfig()
			assert.Error(t, err)
		})

		t.Run("Valid secret with default expiry", func(t *testing.T) {
			t.Setenv("JWT_SECRET", "test-secret")

			cfg, err := loadJWTConfig()
			assert.NoError(t, err)
			assert.Equal(t, "test-secret", cfg.Secret)
			assert.Equal(t, "1h0m0s", cfg.Expiry.String())
		})
	})

	t.Run("Test loadBootstrapAdminConfig", func(t *testing.T) {
		t.Run("Both unset is valid and disabled", func(t *testing.T) {
			cfg, err := loadBootstrapAdminConfig()
			assert.NoError(t, err)
			assert.False(t, cfg.Enabled())
		})

		t.Run("Both set is valid and enabled", func(t *testing.T) {
			t.Setenv("BOOTSTRAP_ADMIN_EMAIL", "admin@example.com")
			t.Setenv("BOOTSTRAP_ADMIN_PASSWORD", "secret")

			cfg, err := loadBootstrapAdminConfig()
			assert.NoError(t, err)
			assert.True(t, cfg.Enabled())
		})

		t.Run("Only one set fails", func(t *testing.T) {
			t.Setenv("BOOTSTRAP_ADMIN_EMAIL", "admin@example.com")

			_, err := loadBootstrapAdminConfig()
			assert.Error(t, err)
		})
	})
}
