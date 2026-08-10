package command

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/FranciscoHonorat/ordemflow/services/order-service/application/port/input"
	"github.com/FranciscoHonorat/ordemflow/services/order-service/application/port/output"
	"github.com/FranciscoHonorat/ordemflow/services/order-service/domain/entity"
	"github.com/FranciscoHonorat/ordemflow/services/order-service/domain/valueobject"
)

type PlaceOrderItem struct {
	ProductID      string
	Quantity       int64
	UnitPriceCents int64
	Currency       string
}

type PlaceOrderCommand struct {
	CustomerID string
	Items      []PlaceOrderItem
}

type PlaceOrderResult struct {
	OrderID string
}

type PlaceOrderHandler struct {
	uow   output.UnitOfWork
	clock output.Clock
}

func NewPlaceOrderHandler(uow output.UnitOfWork, clock output.Clock) *PlaceOrderHandler {
	return &PlaceOrderHandler{
		uow:   uow,
		clock: clock,
	}
}

func (h *PlaceOrderHandler) Handle(ctx context.Context, cmd PlaceOrderCommand) (PlaceOrderResult, error) {
	now := h.clock.Now()
	if len(cmd.Items) == 0 {
		return PlaceOrderResult{}, fmt.Errorf("place order: order must contain at least one item")
	}
	id, err := uuid.Parse(cmd.CustomerID)
	if err != nil {
		return PlaceOrderResult{}, err
	}
	customerID, err := valueobject.NewCustomerID(id)
	if err != nil {
		return PlaceOrderResult{}, err
	}

	items := make([]valueobject.OrderItem, 0, len(cmd.Items))
	for i, in := range cmd.Items {
		price, err := valueobject.NewMoney(in.UnitPriceCents, in.Currency)
		if err != nil {
			return PlaceOrderResult{}, fmt.Errorf("place order: item %d: %w", i, err)
		}

		productUUID, err := uuid.Parse(in.ProductID)
		if err != nil {
			return PlaceOrderResult{}, fmt.Errorf("place order: item %d: %w", i, err)
		}

		productID, err := valueobject.NewProductID(productUUID)
		if err != nil {
			return PlaceOrderResult{}, fmt.Errorf("place order: item %d: %w", i, err)
		}

		quantity, err := valueobject.NewQuantity(in.Quantity)
		if err != nil {
			return PlaceOrderResult{}, fmt.Errorf("place order: item %d: %w", i, err)
		}

		item, err := valueobject.NewOrderItem(productID, price, quantity)
		if err != nil {
			return PlaceOrderResult{}, fmt.Errorf("place order: item %d: %w", i, err)
		}

		items = append(items, item)
	}

	orderID, err := valueobject.NewOrderID(uuid.New())
	if err != nil {
		return PlaceOrderResult{}, err
	}
	order, err := entity.NewOrder(orderID, customerID)
	if err != nil {
		return PlaceOrderResult{}, err
	}

	for _, item := range items {
		if err := order.AddItem(item); err != nil {
			return PlaceOrderResult{}, err
		}
	}

	if err := order.Place(now); err != nil {
		return PlaceOrderResult{}, fmt.Errorf("place order %w", err)
	}

	err = h.uow.Do(ctx, func(store output.RepositoryProvider) error {
		if err := store.OrderRepository().Save(ctx, order); err != nil {
			return fmt.Errorf("failed to save order: %w", err)
		}

		if events := order.PullEvents(); len(events) > 0 {
			if err := store.OutboxRepository().SaveEvents(ctx, events); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return PlaceOrderResult{}, err
	}

	return PlaceOrderResult{OrderID: order.OrderID().String()}, nil
}

func (h *PlaceOrderHandler) Execute(ctx context.Context, input input.CreateOrderInput) (uuid.UUID, error) {
	cmdItems := make([]PlaceOrderItem, len(input.Items))
	for i, item := range input.Items {
		cmdItems[i] = PlaceOrderItem{
			ProductID:      item.ProductID.String(),
			Quantity:       int64(item.Quantity),
			UnitPriceCents: item.Price,
			Currency:       "USD", // Assuming USD for simplicity; adjust as needed
		}
	}

	cmd := PlaceOrderCommand{
		CustomerID: input.CustomerID.String(),
		Items:      cmdItems,
	}

	result, err := h.Handle(ctx, cmd)
	if err != nil {
		return uuid.Nil, err
	}

	createdUUID, err := uuid.Parse(result.OrderID)
	if err != nil {
		return uuid.Nil, err
	}

	return createdUUID, nil
}
