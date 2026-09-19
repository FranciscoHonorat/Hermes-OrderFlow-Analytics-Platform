package config

import "fmt"

type BootstrapAdminConfig struct {
	Email    string
	Password string
}

func (c BootstrapAdminConfig) Enabled() bool {
	return c.Email != "" && c.Password != ""
}

func loadBootstrapAdminConfig() (BootstrapAdminConfig, error) {
	email := getEnv("BOOTSTRAP_ADMIN_EMAIL", "")
	password := getEnv("BOOTSTRAP_ADMIN_PASSWORD", "")

	if (email == "") != (password == "") {
		return BootstrapAdminConfig{}, fmt.Errorf("BOOTSTRAP_ADMIN_EMAIL and BOOTSTRAP_ADMIN_PASSWORD must be set together")
	}

	return BootstrapAdminConfig{Email: email, Password: password}, nil
}
