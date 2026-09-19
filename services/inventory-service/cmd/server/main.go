package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/FranciscoHonorat/ordemflow/services/inventory-service/internal/application/command"
	"github.com/FranciscoHonorat/ordemflow/services/inventory-service/internal/config"
	"github.com/FranciscoHonorat/ordemflow/services/inventory-service/internal/infrastructure/http/handler"
	"github.com/FranciscoHonorat/ordemflow/services/inventory-service/internal/infrastructure/http/middleware"
	"github.com/FranciscoHonorat/ordemflow/services/inventory-service/internal/infrastructure/persistence"
	"github.com/FranciscoHonorat/ordemflow/services/inventory-service/internal/infrastructure/persistence/postgres"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Get()
	log.Printf("Starting inventory-service with config: %+v", cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	db, err := postgres.NewConnection(ctx, dsn)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer db.Close()
	log.Println("Database connection established successfully.")

	uow := postgres.NewUnitOfWork(db)
	clock := persistence.NewRealClock()

	reserveHandler := command.NewReserveStockHandler(uow, clock)
	releaseHandler := command.NewReleaseStockHandler(uow, clock)
	replenishHandler := command.NewReplenishStockHandler(uow, clock)

	reserveUC := &reserveStockUseCaseAdapter{handler: reserveHandler}
	releaseUC := &releaseStockUseCaseAdapter{handler: releaseHandler}
	replenishUC := &replenishStockUseCaseAdapter{handler: replenishHandler}

	stockQueries := postgres.NewStockQueries(db)

	stockHandler := handler.NewStockHandler(stockQueries, reserveUC, releaseUC, replenishUC)
	healthHandler := handler.NewHealthHandler()

	r := gin.New()

	r.Use(middleware.RequestID())
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.GET("/health", gin.WrapF(healthHandler.HealthCheck))

	stockGroup := r.Group("/stock")
	{
		stockGroup.POST("/reserve", stockHandler.ReserveStock)
		stockGroup.POST("/release", stockHandler.ReleaseStock)
		stockGroup.POST("/replenish", stockHandler.ReplenishStock)
		stockGroup.GET("/low", stockHandler.ListLowStock)
		stockGroup.GET("/:sku", stockHandler.GetStockBySKU)
	}

	serverAddr := ":" + cfg.HTTP.Port
	if cfg.HTTP.Port == "" {
		serverAddr = ":8081"
	}

	log.Printf("Starting HTTP server on %s", serverAddr)
	if err := http.ListenAndServe(serverAddr, r); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}
