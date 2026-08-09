package valueobject_test

import (
	"testing"

	domainErrors "github.com/FranciscoHonorat/ordemflow/services/order-service/domain/domain-errors"
	customer "github.com/FranciscoHonorat/ordemflow/services/order-service/domain/valueobject"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCustomerID(t *testing.T) {
	t.Run("Test for NewCustomerID", func(t *testing.T) {
		tests := []struct {
			name          string
			id            uuid.UUID
			expectedError error
		}{
			{"Valid id", uuid.New(), nil},
			{"Invalid id", uuid.Nil, domainErrors.ErrInvalidCustomerID},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				p, err := customer.NewCustomerID(tt.id)

				assert.ErrorIs(t, err, tt.expectedError)

				if tt.expectedError == nil {
					require.NotNil(t, p)
					assert.Equal(t, tt.id, p.ID())
				}
			})
		}
	})

	t.Run("Test for String method", func(t *testing.T) {
		id := uuid.New()
		c, err := customer.NewCustomerID(id)

		require.NoError(t, err)
		assert.Equal(t, id.String(), c.String())
	})

	t.Run("Test for Equal method", func(t *testing.T) {
		id1 := uuid.New()
		id2 := uuid.New()

		tests := []struct {
			name     string
			p1       customer.CustomerID
			p2       customer.CustomerID
			expected bool
		}{
			{"Equal ID", customer.NewCustomerIDMust(id1), customer.NewCustomerIDMust(id1), true},
			{"Diferente ID", customer.NewCustomerIDMust(id1), customer.NewCustomerIDMust(id2), false},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.expected, tt.p1.Equal(tt.p2))
			})
		}
	})

	t.Run("Test for MarhsalJSON", func(t *testing.T) {
		id1 := uuid.New()
		tests := []struct {
			name     string
			c        customer.CustomerID
			expected string
		}{
			{"ID", customer.NewCustomerIDMust(id1), `{"id":"` + id1.String() + `"}`},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				jsonData, err := tt.c.MarshalJSON()
				require.NoError(t, err)
				assert.JSONEq(t, tt.expected, string(jsonData))
			})
		}
	})

	t.Run("Test for UnmarshalJSON", func(t *testing.T) {
		id1 := uuid.New()
		tests := []struct {
			name          string
			inputJSON     string
			expectedID    uuid.UUID
			expectedError error
		}{
			{"Valid JSON", `{"id":"` + id1.String() + `"}`, id1, nil},
			{"Invalid JSON", `{"id":"00000000-0000-0000-0000-000000000000"}`, uuid.Nil, domainErrors.ErrInvalidCustomerID},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var c customer.CustomerID
				err := c.UnmarshalJSON([]byte(tt.inputJSON))

				assert.ErrorIs(t, err, tt.expectedError)

				if tt.expectedError == nil {
					assert.Equal(t, tt.expectedID, c.ID())
				}
			})
		}
	})

}
