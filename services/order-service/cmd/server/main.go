package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/FranciscoHonorat/ordemflow/services/order-service/config"
	"github.com/FranciscoHonorat/ordemflow/services/order-service/infrastructure/http/middleware"
	"github.com/FranciscoHonorat/ordemflow/services/order-service/infrastructure/persistence/postgres"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Get()
	log.Printf("Starting order-service with config: %+v", cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := postgres.NewConnection(ctx, cfg.Database.Host+":"+cfg.Database.Port+" user="+cfg.Database.User+" password="+cfg.Database.Password+" dbname="+cfg.Database.Name+" sslmode="+cfg.Database.SSLMode)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer db.Close()
	log.Println("Database connection established successfully.")

	orderMapper := postgres.NewOrderMapper()

	orderRepo := postgres.NewOrderRepository(db, orderMapper)
	outboxRepo := postgres.NewOutboxRepository(db)
	uow := postgres.NewUnitOfWork(db)

	_ = orderRepo
	_ = outboxRepo
	_ = uow

	if cfg.Env == "development" {
		log.Println("Running in development mode. Performing database migrations...")
	}

	r := gin.New()

	r.Use(middleware.RequestID())

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "time": time.Now().UTC()})
	})

	serverAddr := ":" + cfg.HTTP.Port
	if cfg.HTTP.Port == "" {
		serverAddr = ":8080"
	}

	log.Printf("Starting HTTP server on %s", serverAddr)
	if err := r.Run(serverAddr); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}
