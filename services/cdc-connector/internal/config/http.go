package config

type HTTPServerConfig struct {
	Port string
}

func loadHTTPConfig() (*HTTPServerConfig, error) {
	port := getEnv("HTTP_PORT", "8082")

	return &HTTPServerConfig{
		Port: port,
	}, nil
}
