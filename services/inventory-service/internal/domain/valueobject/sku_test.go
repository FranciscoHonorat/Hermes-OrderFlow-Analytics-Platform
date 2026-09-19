package valueobject_test

import (
	"testing"

	domainErrors "github.com/FranciscoHonorat/ordemflow/services/inventory-service/internal/domain/domain-errors"
	"github.com/FranciscoHonorat/ordemflow/services/inventory-service/internal/domain/valueobject"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSKU(t *testing.T) {
	t.Run("Test NewSKU method", func(t *testing.T) {
		tests := []struct {
			name          string
			value         string
			expectedValue string
			expectedError error
		}{
			{
				name:          "Valid SKU",
				value:         "sku-1",
				expectedValue: "sku-1",
				expectedError: nil,
			},
			{
				name:          "Trims surrounding whitespace",
				value:         "  sku-1  ",
				expectedValue: "sku-1",
				expectedError: nil,
			},
			{
				name:          "Invalid SKU (empty)",
				value:         "",
				expectedError: domainErrors.ErrInvalidSKU,
			},
			{
				name:          "Invalid SKU (only whitespace)",
				value:         "   ",
				expectedError: domainErrors.ErrInvalidSKU,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				sku, err := valueobject.NewSKU(tt.value)

				if tt.expectedError != nil {
					assert.ErrorIs(t, err, tt.expectedError)
					assert.True(t, sku.IsZero())
				} else {
					require.NoError(t, err)
					assert.Equal(t, tt.expectedValue, sku.String())
					assert.False(t, sku.IsZero())
				}
			})
		}
	})

	t.Run("Test IsZero method", func(t *testing.T) {
		var zero valueobject.SKU
		assert.True(t, zero.IsZero())

		sku, err := valueobject.NewSKU("sku-1")
		require.NoError(t, err)
		assert.False(t, sku.IsZero())
	})

	t.Run("Test Equal method", func(t *testing.T) {
		sku1, err := valueobject.NewSKU("sku-1")
		require.NoError(t, err)

		sku2, err := valueobject.NewSKU("sku-1")
		require.NoError(t, err)

		sku3, err := valueobject.NewSKU("sku-2")
		require.NoError(t, err)

		assert.True(t, sku1.Equal(sku2))
		assert.False(t, sku1.Equal(sku3))
	})

	t.Run("Test MarshalJSON", func(t *testing.T) {
		sku, err := valueobject.NewSKU("sku-1")
		require.NoError(t, err)

		jsonData, err := sku.MarshalJSON()
		require.NoError(t, err)
		assert.JSONEq(t, `"sku-1"`, string(jsonData))
	})

	t.Run("Test UnmarshalJSON", func(t *testing.T) {
		tests := []struct {
			name          string
			inputJSON     string
			expectedValue string
			expectedError error
		}{
			{"Valid JSON string", `"sku-1"`, "sku-1", nil},
			{"Invalid JSON empty string", `""`, "", domainErrors.ErrInvalidSKU},
			{"Malformed syntax", `{"sku": "1"}`, "", assert.AnError},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var sku valueobject.SKU
				err := sku.UnmarshalJSON([]byte(tt.inputJSON))

				if tt.expectedError != nil {
					if tt.expectedError == domainErrors.ErrInvalidSKU {
						assert.ErrorIs(t, err, domainErrors.ErrInvalidSKU)
					} else {
						assert.Error(t, err)
					}
				} else {
					require.NoError(t, err)
					assert.Equal(t, tt.expectedValue, sku.String())
				}
			})
		}
	})
}
