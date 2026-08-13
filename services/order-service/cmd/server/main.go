package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/FranciscoHonorat/ordemflow/services/order-service/application/command"
	"github.com/FranciscoHonorat/ordemflow/services/order-service/config"
	"github.com/FranciscoHonorat/ordemflow/services/order-service/infrastructure/http/handler"
	"github.com/FranciscoHonorat/ordemflow/services/order-service/infrastructure/http/middleware"
	"github.com/FranciscoHonorat/ordemflow/services/order-service/infrastructure/persistence"
	"github.com/FranciscoHonorat/ordemflow/services/order-service/infrastructure/persistence/postgres"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Get()
	log.Printf("Starting order-service with config: %+v", cfg)

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

	orderMapper := postgres.NewOrderMapper()

	orderRepo := postgres.NewOrderRepository(db, orderMapper)
	outboxRepo := postgres.NewOutboxRepository(db)
	uow := postgres.NewUnitOfWork(db)

	_ = orderRepo
	_ = outboxRepo
	_ = uow

	clock := persistence.NewRealClock()

	placeHandler := command.NewPlaceOrderHandler(uow, clock)
	confirmHandler := command.NewConfirmOrderHandler(uow, clock)
	cancelHandler := command.NewCancelOrderHandler(uow, clock)
	shipHandler := command.NewShipOrderHandler(uow, clock)

	placeUC := &placeOrderUseCaseAdapter{handler: placeHandler}
	confirmUC := &confirmOrderUseCaseAdapter{handler: confirmHandler}
	cancelUC := &cancelOrderUseCaseAdapter{handler: cancelHandler}
	shipUC := &shipOrderUseCaseAdapter{handler: shipHandler}

	queries := &queriesOrderUseCaseAdapter{
		queries: postgres.NewOrderQueries(db),
	}

	orderHandler := handler.NewOrderHandler(
		queries,
		placeUC,
		confirmUC,
		cancelUC,
		shipUC,
	)

	r := gin.New()

	r.Use(middleware.RequestID())

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "time": time.Now().UTC()})
	})

	orderRoutes := r.Group("/orders")
	{
		orderRoutes.POST("", orderHandler.PlaceOrder)
		orderRoutes.POST("/confirm", orderHandler.ConfirmOrder)
		orderRoutes.POST("/cancel", orderHandler.CancelOrder)
		orderRoutes.POST("/ship", orderHandler.ShipOrder)
		orderRoutes.GET("/:id", orderHandler.GetOrderByID)
	}

	serverAddr := ":" + cfg.HTTP.Port
	if cfg.HTTP.Port == "" {
		serverAddr = ":8080"
	}

	log.Printf("Starting HTTP server on %s", serverAddr)
	if err := r.Run(serverAddr); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}
