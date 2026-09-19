package config

import (
	"strings"
)

type MessagingConfig struct {
	Brokers  []string
	ClientID string
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

	clientID := getEnv("MESSAGING_CLIENT_ID", "cdc-connector")

	return MessagingConfig{
		Brokers:  brokers,
		ClientID: clientID,
	}, nil
}
