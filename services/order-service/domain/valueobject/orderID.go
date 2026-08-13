package valueobject

import (
	"encoding/json"

	domainErrors "github.com/FranciscoHonorat/ordemflow/services/order-service/domain/domain-errors"
	"github.com/google/uuid"
)

type OrderID struct {
	id uuid.UUID
}

func NewOrderID(id uuid.UUID) (OrderID, error) {
	if id == uuid.Nil {
		return OrderID{}, domainErrors.ErrInvalidOrderID
	}

	return OrderID{id: id}, nil
}

func NewOrderIDMust(id uuid.UUID) OrderID {
	orderID, err := NewOrderID(id)
	if err != nil {
		panic(err)
	}
	return orderID
}

func ParseOrderID(idStr string) (OrderID, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return OrderID{}, domainErrors.ErrInvalidOrderID
	}

	return NewOrderID(id)
}

func (o OrderID) ID() uuid.UUID {
	return o.id
}

func (o OrderID) String() string {
	return o.id.String()
}

func (o OrderID) Equal(other OrderID) bool {
	return o.id == other.id
}

func (o OrderID) IsZero() bool {
	return o.id == uuid.Nil
}

func (o OrderID) MarshalJSON() ([]byte, error) {
	auxOrder := struct {
		ID string `json:"id"`
	}{
		ID: o.id.String(),
	}
	return json.Marshal(auxOrder)
}

func (o *OrderID) UnmarshalJSON(data []byte) error {
	var order struct {
		ID uuid.UUID
	}

	if err := json.Unmarshal(data, &order); err != nil {
		return err
	}

	if order.ID == uuid.Nil {
		return domainErrors.ErrInvalidOrderID
	}

	o.id = order.ID
	return nil
}
