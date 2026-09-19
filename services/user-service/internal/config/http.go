package config

type HTTPServerConfig struct {
	Port string
	Mode string
}

func loadHTTPConfig() (*HTTPServerConfig, error) {
	port := getEnv("HTTP_PORT", "8083")
	mode := getEnv("HTTP_MODE", "release")

	return &HTTPServerConfig{
		Port: port,
		Mode: mode,
	}, nil
}
