package valueobject

import (
	"encoding/json"
	"strings"

	domainErrors "github.com/FranciscoHonorat/ordemflow/services/inventory-service/internal/domain/domain-errors"
)

// SKU é o identificador de negócio de um item de estoque (Stock Keeping Unit).
type SKU struct {
	value string
}

func NewSKU(value string) (SKU, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return SKU{}, domainErrors.ErrInvalidSKU
	}
	return SKU{value: trimmed}, nil
}

func NewSKUMust(value string) SKU {
	sku, err := NewSKU(value)
	if err != nil {
		panic(err)
	}
	return sku
}

func (s SKU) String() string {
	return s.value
}

func (s SKU) IsZero() bool {
	return s.value == ""
}

func (s SKU) Equal(o SKU) bool {
	return s.value == o.value
}

func (s SKU) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.value)
}

func (s *SKU) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	sku, err := NewSKU(value)
	if err != nil {
		return err
	}

	*s = sku
	return nil
}
