package main

import (
	"context"

	"github.com/FranciscoHonorat/ordemflow/services/inventory-service/internal/application/command"
	"github.com/FranciscoHonorat/ordemflow/services/inventory-service/internal/application/port/input"
)

type reserveStockUseCaseAdapter struct {
	handler *command.ReserveStockHandler
}

func (a *reserveStockUseCaseAdapter) Execute(ctx context.Context, in input.ReserveStockInput) error {
	return a.handler.Execute(ctx, in)
}

type releaseStockUseCaseAdapter struct {
	handler *command.ReleaseStockHandler
}

func (a *releaseStockUseCaseAdapter) Execute(ctx context.Context, in input.ReleaseStockInput) error {
	return a.handler.Execute(ctx, in)
}

type replenishStockUseCaseAdapter struct {
	handler *command.ReplenishStockHandler
}

func (a *replenishStockUseCaseAdapter) Execute(ctx context.Context, in input.ReplenishStockInput) error {
	return a.handler.Execute(ctx, in)
}
