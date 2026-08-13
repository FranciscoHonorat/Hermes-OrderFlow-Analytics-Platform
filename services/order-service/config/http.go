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
		orderGroup.POST("", gin.WrapF(orderHandler.PlaceOrder))
		orderGroup.GET("", gin.WrapF(orderHandler.ListOrders))

		orderGroup.POST("/confirm", gin.WrapF(orderHandler.ConfirmOrder))
		orderGroup.POST("/cancel", gin.WrapF(orderHandler.CancelOrder))
		orderGroup.POST("/ship", gin.WrapF(orderHandler.ShipOrder))

		orderGroup.GET("/:order_id", gin.WrapF(orderHandler.GetOrderByID))
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
