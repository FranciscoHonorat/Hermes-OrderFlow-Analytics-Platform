package event

import (
	"time"

	domainErrors "github.com/FranciscoHonorat/ordemflow/services/order-service/domain/domain-errors"
	"github.com/FranciscoHonorat/ordemflow/services/order-service/domain/valueobject"
	"github.com/FranciscoHonorat/ordemflow/shared/events"
)

var _ events.DomainEvent = (*OrderAdded)(nil)

type OrderAdded struct {
	events.BaseEvent
	OrderID    valueobject.OrderID
	ProductID  valueobject.ProductID
	Quantity   valueobject.Quantity
	UnitPrice  valueobject.Money
	TotalPrice valueobject.Money
}

func NewOrderAdded(orderID valueobject.OrderID, productID valueobject.ProductID, quantity valueobject.Quantity, unitPrice, totalPrice valueobject.Money, occurredAt time.Time) OrderAdded {
	return OrderAdded{
		BaseEvent:  events.NewBaseEvent("order.item_added", orderID.String(), occurredAt),
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
