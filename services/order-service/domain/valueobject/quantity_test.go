package valueobject_test

import (
	"testing"

	domainErrors "github.com/FranciscoHonorat/ordemflow/services/order-service/domain/domain-errors"
	quantity "github.com/FranciscoHonorat/ordemflow/services/order-service/domain/valueobject"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuantity(t *testing.T) {
	t.Run("Test NewQuantity method", func(t *testing.T) {
		tests := []struct {
			name          string
			value         int64
			expectedError error
		}{
			{
				name:          "Valid quantity",
				value:         10,
				expectedError: nil,
			},
			{
				name:          "Invalid quantity (zero)",
				value:         0,
				expectedError: domainErrors.ErrInvalidQuantity,
			},
			{
				name:          "Invalid quantity (negative)",
				value:         -5,
				expectedError: domainErrors.ErrInvalidQuantity,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				q, err := quantity.NewQuantity(tt.value)

				if tt.expectedError != nil {
					assert.ErrorIs(t, err, tt.expectedError)
				} else {
					require.NoError(t, err)
					require.NotNil(t, q)
					assert.Equal(t, tt.value, q.Value())
				}
			})
		}
	})

	t.Run("Test Equal method", func(t *testing.T) {
		q1, err := quantity.NewQuantity(10)
		require.NoError(t, err)

		q2, err := quantity.NewQuantity(10)
		require.NoError(t, err)

		q3, err := quantity.NewQuantity(5)
		require.NoError(t, err)

		assert.True(t, q1.Equal(q2), "Expected q1 to be equal to q2")
		assert.False(t, q1.Equal(q3), "Expected q1 to not be equal to q3")
	})

	t.Run("Test MarshalJSON", func(t *testing.T) {
		q, err := quantity.NewQuantity(10)
		require.NoError(t, err)

		tests := []struct {
			name     string
			quantity quantity.Quantity
			expected string
		}{
			{"Valid quantity serialization", q, `10`},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				jsonData, err := tt.quantity.MarshalJSON()
				require.NoError(t, err)
				assert.JSONEq(t, tt.expected, string(jsonData))
			})
		}
	})

	t.Run("Test UnmarshalJSON", func(t *testing.T) {
		tests := []struct {
			name          string
			inputJSON     string
			expectedValue int64
			expectedError error
		}{
			{"Valid JSON scalar", `10`, 10, nil},
			{"Invalid JSON zero value", `0`, 0, domainErrors.ErrInvalidQuantity},
			{"Invalid JSON negative value", `-5`, 0, domainErrors.ErrInvalidQuantity},
			{"Malformed syntax", `{"qty": 10}`, 0, assert.AnError},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var unmarshaledQuantity quantity.Quantity
				err := unmarshaledQuantity.UnmarshalJSON([]byte(tt.inputJSON))

				if tt.expectedError != nil {
					if tt.expectedError == domainErrors.ErrInvalidQuantity {
						assert.ErrorIs(t, err, domainErrors.ErrInvalidQuantity)
					} else {
						assert.Error(t, err)
					}
				} else {
					require.NoError(t, err)
					assert.Equal(t, tt.expectedValue, unmarshaledQuantity.Value())
				}
			})
		}
	})
}
