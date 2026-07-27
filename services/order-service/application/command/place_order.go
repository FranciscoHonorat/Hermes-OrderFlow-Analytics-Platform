package command

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/FranciscoHonorat/ordemflow/services/order-service/domain/entity"
	"github.com/FranciscoHonorat/ordemflow/services/order-service/domain/repository"
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
	orders repository.OrderRepository
}

func NewPlaceOrderHandler(orders repository.OrderRepository) *PlaceOrderHandler {
	return &PlaceOrderHandler{
		orders: orders,
	}
}

func (h *PlaceOrderHandler) Handle(ctx context.Context, cmd PlaceOrderCommand) (PlaceOrderResult, error) {
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

		quantity, err := valueobject.NewQuantity(in.Quantity)

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

	if err := h.orders.Save(ctx, order); err != nil {
		return PlaceOrderResult{}, err
	}

	return PlaceOrderResult{OrderID: order.OrderID().String()}, nil
}
