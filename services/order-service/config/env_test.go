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
}
