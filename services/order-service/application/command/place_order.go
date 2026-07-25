package Command

import (
aggregate "github.com/FranciscoHonorat/ordemflow/services/order-service/domain/entity"
"github.com/FranciscoHonorat/ordemflow/services/order-service/domain/repository"
)

type PlaceOrderItem struct {
ProductID string
Quantity int
UnitPriceCents int64
Currency string
}

type PlaceOrderCommand struct {
CustomerID string
Items []PlaceOrderItem
}

type PlaceOrderResult struct {
OrderID string
}

type PlaceOrderHandler interface {
orders repository.OrderRepository
}

func NewPlaceOrderHandler(orders repository.OrderRepository) *PlaceOrderHandler {
return &PlaceOrderHandler{
        orders: orders,
    }
}

func (h *PlaceOrderHandle) Handle(ctx context.Context, cmd PlaceOrderCommand) (PlaceOrderResult, error) {
	id, err := uuid.Parse(cmd.CustomerID)
	if err != nil {
		return PlaceOrderResult{}, err
	}
	customerID, err := valueobject.NewCustomerID(id)
	if err != nil {
		return PlaceOrderResult{}, err
	}

	item := make([]valueobject.OrderItem.len(cmd.Items))

	for i, item range cmd.Items{
		products[i] = valueobject.OrderItem{
			ProductID: item.ProductID,
			Quantity: item.Quantity,
			UnitPriceCents: item.UnitPrice,
			Currency: item.Currency,
		}
	}
}