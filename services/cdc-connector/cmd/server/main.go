package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/FranciscoHonorat/ordemflow/services/cdc-connector/internal/config"
	"github.com/FranciscoHonorat/ordemflow/services/cdc-connector/internal/relay"
	"github.com/FranciscoHonorat/ordemflow/shared/messaging/kafka"
	"github.com/FranciscoHonorat/ordemflow/shared/outbox"
	sharedpg "github.com/FranciscoHonorat/ordemflow/shared/postgres"
)

func main() {
	cfg := config.Get()
	log.Printf("Starting cdc-connector with config: %+v", cfg)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	sources := make([]relay.Source, 0, len(cfg.Sources))
	for _, sourceCfg := range cfg.Sources {
		connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		db, err := sharedpg.NewConnection(connectCtx, cfg.Database.DSN(sourceCfg.DBName))
		cancel()
		if err != nil {
			log.Fatalf("Failed to connect to database %q for source %q: %v", sourceCfg.DBName, sourceCfg.Name, err)
		}
		defer db.Close()

		sources = append(sources, relay.Source{
			Name:   sourceCfg.Name,
			Outbox: outbox.NewPostgresRepository(db.Pool),
		})
		log.Printf("Connected source %q (database %q)", sourceCfg.Name, sourceCfg.DBName)
	}

	producer, err := kafka.NewProducer(cfg.Messaging.Brokers, cfg.Messaging.ClientID)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer producer.Close()

	r := relay.New(producer, cfg.Relay.BatchSize, sources...)

	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/health", func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status":    "healthy",
				"timestamp": time.Now().UTC(),
				"system":    "cdc-connector",
			})
		})

		addr := ":" + cfg.HTTP.Port
		log.Printf("Starting HTTP server on %s", addr)
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	log.Printf("Starting relay loop, polling every %s in batches of %d", cfg.Relay.PollInterval, cfg.Relay.BatchSize)
	r.Run(ctx, cfg.Relay.PollInterval)
	log.Println("Shutting down cdc-connector")
}
