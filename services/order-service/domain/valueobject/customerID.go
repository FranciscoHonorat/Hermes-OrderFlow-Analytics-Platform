package valueobject

import (
	"encoding/json"

	domainErrors "github.com/FranciscoHonorat/ordemflow/services/order-service/domain/domain-errors"
	"github.com/google/uuid"
)

type CustomerID struct {
	id uuid.UUID
}

func NewCustomerID(id uuid.UUID) (CustomerID, error) {
	if id == uuid.Nil {
		return CustomerID{}, domainErrors.ErrInvalidCustomerID
	}

	return CustomerID{id: id}, nil
}

func NewCustomerIDMust(id uuid.UUID) CustomerID {
	m, err := NewCustomerID(id)
	if err != nil {
		panic(err)
	}
	return m
}

func ParseCustomerID(idStr string) (CustomerID, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return CustomerID{}, domainErrors.ErrInvalidCustomerID
	}

	return NewCustomerID(id)
}

func (c CustomerID) ID() uuid.UUID {
	return c.id
}

func (c CustomerID) String() string {
	return c.id.String()
}

func (c CustomerID) Equal(o CustomerID) bool {
	return c.id == o.id
}

func (c CustomerID) IsZero() bool {
	return c.id == uuid.Nil
}

func (c CustomerID) MarshalJSON() ([]byte, error) {
	auxCustomer := struct {
		ID string `json:"id"`
	}{
		ID: c.id.String(),
	}
	return json.Marshal(auxCustomer)
}

func (c *CustomerID) UnmarshalJSON(data []byte) error {
	var customer struct {
		ID uuid.UUID
	}

	if err := json.Unmarshal(data, &customer); err != nil {
		return err
	}

	if customer.ID == uuid.Nil {
		return domainErrors.ErrInvalidCustomerID
	}

	c.id = customer.ID
	return nil
}
