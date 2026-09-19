package config

import (
	"fmt"
	"strconv"
	"time"
)

type RelayConfig struct {
	PollInterval time.Duration
	BatchSize    int
}

func loadRelayConfig() (RelayConfig, error) {
	intervalStr := getEnv("POLL_INTERVAL", "2s")
	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		return RelayConfig{}, fmt.Errorf("invalid POLL_INTERVAL %q: %w", intervalStr, err)
	}

	batchSizeStr := getEnv("POLL_BATCH_SIZE", "50")
	batchSize, err := strconv.Atoi(batchSizeStr)
	if err != nil || batchSize <= 0 {
		return RelayConfig{}, fmt.Errorf("invalid POLL_BATCH_SIZE %q: must be a positive integer", batchSizeStr)
	}

	return RelayConfig{
		PollInterval: interval,
		BatchSize:    batchSize,
	}, nil
}
