package config

import (
	"context"
	"net/http"
	"time"

	"github.com/FranciscoHonorat/ordemflow/services/order-service/application/port/input"
	"github.com/FranciscoHonorat/ordemflow/services/order-service/infrastructure/http/handler"
	"github.com/gin-gonic/gin"
)

type HTTPServerConfig struct {
	Port string
	Mode string
}

type HTTPServer struct {
	server *http.Server
}

func NewHTTPServer(
	cfg *HTTPServerConfig,
	queries input.OrderQueries,
	place input.PlaceOrderUseCase,
	confirm input.ConfirmOrderUseCase,
	cancel input.CancelOrderUseCase,
	ship input.ShipOrderUseCase,
) *HTTPServer {
	if cfg.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())

	orderHandler := handler.NewOrderHandler(queries, place, confirm, cancel, ship)
	healthHandler := handler.NewHealthHandler()

	router.GET("/health", gin.WrapF(healthHandler.HealthCheck))

	orderGroup := router.Group("/orders")
	{
		orderGroup.POST("", func(c *gin.Context) { orderHandler.PlaceOrder(c) })
		orderGroup.GET("", func(c *gin.Context) { orderHandler.ListOrders(c) })

		orderGroup.POST("/confirm", func(c *gin.Context) { orderHandler.ConfirmOrder(c) })
		orderGroup.POST("/cancel", func(c *gin.Context) { orderHandler.CancelOrder(c) })
		orderGroup.POST("/ship", func(c *gin.Context) { orderHandler.ShipOrder(c) })

		orderGroup.GET("/:order_id", func(c *gin.Context) { orderHandler.GetOrderByID(c) })
	}

	return &HTTPServer{
		server: &http.Server{
			Addr:         ":" + cfg.Port,
			Handler:      router,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
	}
}

func (s *HTTPServer) Start() error {
	return s.server.ListenAndServe()
}

func (s *HTTPServer) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func loadHTTPConfig() (*HTTPServerConfig, error) {
	port := getEnv("HTTP_PORT", "8080")
	mode := getEnv("HTTP_MODE", "release")

	return &HTTPServerConfig{
		Port: port,
		Mode: mode,
	}, nil
}
