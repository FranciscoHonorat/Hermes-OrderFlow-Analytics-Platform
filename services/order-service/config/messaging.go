package config

import (
	"strings"
	"time"
)

type MessagingConfig struct {
	BrokerRaw    string
	Brokers      []string
	ClientID     string
	WriteTimeout time.Duration
}

func loadMessagingConfig() (MessagingConfig, error) {
	brokerRaw, err := getRequiredEnv("MESSAGING_BROKER")
	if err != nil {
		return MessagingConfig{}, err
	}

	rawList := strings.Split(brokerRaw, ",")
	brokers := make([]string, 0, len(rawList))

	for _, broker := range rawList {
		broker = strings.TrimSpace(broker)
		if broker != "" {
			brokers = append(brokers, broker)
		}
	}

	clientID := getEnv("MESSAGING_CLIENT_ID", "order-service")

	writeTimeoutStr := getEnv("MESSAGING_WRITE_TIMEOUT", "5s")
	writeTimeout, err := time.ParseDuration(writeTimeoutStr)
	if err != nil {
		return MessagingConfig{}, err
	}

	return MessagingConfig{
		BrokerRaw:    brokerRaw,
		Brokers:      brokers,
		ClientID:     clientID,
		WriteTimeout: writeTimeout,
	}, nil
}
