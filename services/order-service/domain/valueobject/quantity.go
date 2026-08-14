package valueobject

import (
	"encoding/json"

	domainErrors "github.com/FranciscoHonorat/ordemflow/services/order-service/domain/domain-errors"
)

type Quantity struct {
	value int64
}

func NewQuantity(value int64) (Quantity, error) {
	if value <= 0 {
		return Quantity{}, domainErrors.ErrInvalidQuantity
	}
	return Quantity{value: value}, nil
}

func NewQuantityMust(value int64) Quantity {
	q, err := NewQuantity(value)
	if err != nil {
		panic(err)
	}
	return q
}

func (q Quantity) Validate() error {
	if q.value <= 0 {
		return domainErrors.ErrInvalidQuantity
	}
	return nil
}

func (q Quantity) Value() int64 {
	return q.value
}

func (q Quantity) Equal(o Quantity) bool {
	return q.value == o.value
}

func (q Quantity) MarshalJSON() ([]byte, error) {
	return json.Marshal(q.value)
}

func (q *Quantity) UnmarshalJSON(data []byte) error {
	var value int64

	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	if value <= 0 {
		return domainErrors.ErrInvalidQuantity
	}

	q.value = value
	return nil
}
