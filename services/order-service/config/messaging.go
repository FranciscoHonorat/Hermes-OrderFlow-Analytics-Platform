package config

import "time"

type MessagingConfig struct {
	Brokers     string
	ClientID    string
	WriteTimout time.Duration
}

type MessagingSystem struct {
	publisher *kafka.EventPublisher
	producer  *kafka.EventProducer
}
