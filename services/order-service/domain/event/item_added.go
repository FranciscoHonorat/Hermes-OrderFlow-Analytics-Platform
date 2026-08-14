package event

import (
	"time"

	domainErrors "github.com/FranciscoHonorat/ordemflow/services/order-service/domain/domain-errors"
	"github.com/FranciscoHonorat/ordemflow/services/order-service/domain/valueobject"
)

var _ DomainEvent = (*OrderPlaced)(nil)

type OrderAdded struct {
	BaseEvent
	OrderID    valueobject.OrderID
	ProductID  valueobject.ProductID
	Quantity   valueobject.Quantity
	UnitPrice  valueobject.Money
	TotalPrice valueobject.Money
}

func NewOrderAdded(orderID valueobject.OrderID, productID valueobject.ProductID, quantity valueobject.Quantity, unitPrice, totalPrice valueobject.Money, occurredAt time.Time) OrderAdded {
	return OrderAdded{
		BaseEvent:  NewBaseEvent("order.item_added", orderID.String(), occurredAt),
		OrderID:    orderID,
		ProductID:  productID,
		Quantity:   quantity,
		UnitPrice:  unitPrice,
		TotalPrice: totalPrice,
	}
}

func (event OrderAdded) ValidateOrderAdded() error {
	if event.OrderID.IsZero() {
		return domainErrors.ErrInvalidOrderID
	}
	if event.ProductID.IsZero() {
		return domainErrors.ErrInvalidProductID
	}
	if err := event.Quantity.Validate(); err != nil {
		return err
	}
	if err := event.UnitPrice.Validate(); err != nil {
		return err
	}
	if err := event.TotalPrice.Validate(); err != nil {
		return err
	}
	return nil
}
