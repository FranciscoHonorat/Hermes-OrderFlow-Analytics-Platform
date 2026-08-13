package main

import (
	"context"

	"github.com/FranciscoHonorat/ordemflow/services/order-service/application/command"
	"github.com/FranciscoHonorat/ordemflow/services/order-service/application/port/input"
	"github.com/google/uuid"
)

type placeOrderUseCaseAdapter struct {
	handler *command.PlaceOrderHandler
}

func (a *placeOrderUseCaseAdapter) Execute(ctx context.Context, in input.CreateOrderInput) (uuid.UUID, error) {
	return a.handler.Execute(ctx, in)
}

type confirmOrderUseCaseAdapter struct {
	handler *command.ConfirmOrderHandler
}

func (a *confirmOrderUseCaseAdapter) Execute(ctx context.Context, orderID uuid.UUID) error {
	return a.handler.Handle(ctx, command.ConfirmOrderCommand{
		OrderID: orderID.String(),
	})
}

type cancelOrderUseCaseAdapter struct {
	handler *command.CancelOrderHandler
}

func (a *cancelOrderUseCaseAdapter) Execute(ctx context.Context, orderID uuid.UUID) error {
	return a.handler.Handle(ctx, command.CancelOrderCommand{
		OrderID: orderID.String(),
	})
}

type shipOrderUseCaseAdapter struct {
	handler *command.ShipOrderHandler
}

func (a *shipOrderUseCaseAdapter) Execute(ctx context.Context, orderID uuid.UUID) error {
	return a.handler.Handle(ctx, command.ShipOrderCommand{
		OrderID: orderID.String(),
	})
}

type queriesOrderUseCaseAdapter struct {
	queries input.OrderQueries
}

func (a *queriesOrderUseCaseAdapter) GetOrderByID(ctx context.Context, orderID uuid.UUID) (*input.OrderDTO, error) {
	return a.queries.GetOrderByID(ctx, orderID)
}

func (a *queriesOrderUseCaseAdapter) ListOrders(ctx context.Context, customerID *uuid.UUID, limit, offset int) ([]input.OrderDTO, error) {
	return a.queries.ListOrders(ctx, customerID, limit, offset)
}
