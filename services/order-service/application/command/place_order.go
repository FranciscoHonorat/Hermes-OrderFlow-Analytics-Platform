package command

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

func (h *PlaceOrderHandler) Handle(ctx context.Context, cmd PlaceOrderCommand) (PlaceOrderResult, error) {
	id, err := uuid.Parse(cmd.CustomerID)
	if err != nil {
		return PlaceOrderResult{}, err
	}
	customerID, err := valueobject.NewCustomerID(id)
	if err != nil {
		return PlaceOrderResult{}, err
	}

	item := make([]valueobject.OrderItem.len(cmd.Items))
	for i, in := range cmd.Items {
		price, err := valueobject.NewMoney(in.UnitPriceCents, in.Currency)
		if err != nil {
			return PlaceOrderResult{}, err
		}

		item, err := valueobject.NewOrderItem(in.ProductID, in.Quantity, price)
		if err != nil {
			return PlaceOrderResult{}, err
		}
	}

	order, err := entity.NewOrder(customerID, items)
	if err != nil {
		return PlaceOrderResult{}, err
	}

	if err := h.orders.Save(ctx, order); err != nil {
		return PlaceOrderResult{}, err
	}

	return PlaceOrderResult{OrderID: order.ID().String()}, nil
}