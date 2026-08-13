package valueobject_test

import (
	"testing"

	domainErrors "github.com/FranciscoHonorat/ordemflow/services/order-service/domain/domain-errors"
	orderid "github.com/FranciscoHonorat/ordemflow/services/order-service/domain/valueobject"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderID(t *testing.T) {
	t.Run("Test for NewOrderID", func(t *testing.T) {
		tests := []struct {
			name    string
			id      uuid.UUID
			wantErr error
		}{
			{"Valid id", uuid.New(), nil},
			{"Invalid id", uuid.Nil, domainErrors.ErrInvalidOrderID},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				p, err := orderid.NewOrderID(tt.id)

				assert.ErrorIs(t, err, tt.wantErr)

				if tt.wantErr == nil {
					require.NotNil(t, p)
					assert.Equal(t, tt.id, p.ID())
				}
			})
		}
	})

	t.Run("Test for ParseOrderID", func(t *testing.T) {
		validUUID := uuid.New()
		tests := []struct {
			name          string
			idStr         string
			expectedID    uuid.UUID
			expectedError error
		}{
			{"Valid UUID", validUUID.String(), validUUID, nil},
			{"Invalid UUID", "invalid-uuid", uuid.Nil, domainErrors.ErrInvalidOrderID},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				orderID, err := orderid.ParseOrderID(tt.idStr)

				assert.ErrorIs(t, err, tt.expectedError)

				if tt.expectedError == nil {
					assert.Equal(t, tt.expectedID, orderID.ID())
				}
			})
		}
	})

	t.Run("Test for String method", func(t *testing.T) {
		id := uuid.New()
		orderIDObj, err := orderid.NewOrderID(id)
		require.NoError(t, err)

		assert.Equal(t, id.String(), orderIDObj.String())
	})

	t.Run("Test for Equal method", func(t *testing.T) {
		id1 := uuid.New()
		id2 := uuid.New()

		orderID1 := orderid.NewOrderIDMust(id1)
		orderID2 := orderid.NewOrderIDMust(id1)
		orderID3 := orderid.NewOrderIDMust(id2)

		assert.True(t, orderID1.Equal(orderID2), "Expected orderID1 to be equal to orderID2")
		assert.False(t, orderID1.Equal(orderID3), "Expected orderID1 to not be equal to orderID3")
	})

	t.Run("Test for MarshalJSON", func(t *testing.T) {
		id := uuid.New()
		tests := []struct {
			name     string
			orderID  orderid.OrderID
			expected string
		}{
			{"Valid ID", orderid.NewOrderIDMust(id), `{"id":"` + id.String() + `"}`},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				jsonData, err := tt.orderID.MarshalJSON()
				require.NoError(t, err)
				assert.JSONEq(t, tt.expected, string(jsonData))
			})
		}
	})

	t.Run("Test for UnmarshalJSON", func(t *testing.T) {
		id := uuid.New()
		tests := []struct {
			name          string
			inputJSON     string
			expectedID    uuid.UUID
			expectedError error
		}{
			{"Valid JSON", `{"id":"` + id.String() + `"}`, id, nil},
			{"Invalid JSON ID", `{"id":"00000000-0000-0000-0000-000000000000"}`, uuid.Nil, domainErrors.ErrInvalidOrderID},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var unmarshalledOrderID orderid.OrderID
				err := unmarshalledOrderID.UnmarshalJSON([]byte(tt.inputJSON))

				assert.ErrorIs(t, err, tt.expectedError)

				if tt.expectedError == nil {
					assert.Equal(t, tt.expectedID, unmarshalledOrderID.ID())
				}
			})
		}
	})
}
