package valueobject_test

import (
	"testing"

	domainErrors "github.com/FranciscoHonorat/ordemflow/services/order-service/domain/domain-errors"
	status "github.com/FranciscoHonorat/ordemflow/services/order-service/domain/valueobject"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderStatus(t *testing.T) {
	t.Run("Test for IsValid method", func(t *testing.T) {
		tests := []struct {
			name   string
			status status.OrderStatus
			want   bool
		}{
			{"Valid status PENDING", status.OrderStatusPending, true},
			{"Valid status CONFIRMED", status.OrderStatusConfirmed, true},
			{"Valid status SHIPPED", status.OrderStatusShipped, true},
			{"Valid status DELIVERED", status.OrderStatusDelivered, true},
			{"Valid status CANCELLED", status.OrderStatusCancelled, true},
			{"Invalid status", "INVALID_STATUS", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.want, tt.status.IsValid())
			})
		}
	})

	t.Run("Test for String method", func(t *testing.T) {
		tests := []struct {
			name   string
			status status.OrderStatus
			want   string
		}{
			{"Status PENDING", status.OrderStatusPending, "PENDING"},
			{"Status CONFIRMED", status.OrderStatusConfirmed, "CONFIRMED"},
			{"Status SHIPPED", status.OrderStatusShipped, "SHIPPED"},
			{"Status DELIVERED", status.OrderStatusDelivered, "DELIVERED"},
			{"Status CANCELLED", status.OrderStatusCancelled, "CANCELLED"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.want, tt.status.String())
			})
		}
	})

	t.Run("Test for MarshalJSON", func(t *testing.T) {
		tests := []struct {
			name     string
			status   status.OrderStatus
			expected string
		}{
			{"Status PENDING", status.OrderStatusPending, `"PENDING"`},
			{"Status CONFIRMED", status.OrderStatusConfirmed, `"CONFIRMED"`},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				jsonData, err := tt.status.MarshalJSON()
				require.NoError(t, err)
				assert.JSONEq(t, tt.expected, string(jsonData))
			})
		}
	})

	t.Run("Test for UnmarshalJSON", func(t *testing.T) {
		tests := []struct {
			name          string
			inputJSON     string
			expected      status.OrderStatus
			expectedError error
		}{
			{"Valid PENDING", `"PENDING"`, status.OrderStatusPending, nil},
			{"Valid CONFIRMED", `"CONFIRMED"`, status.OrderStatusConfirmed, nil},
			{"Invalid Status Value", `"INVALID_STATUS"`, "", domainErrors.ErrInvalidOrderStatus},
			{"Malformed JSON Syntax", `{"status": PENDING}`, "", assert.AnError},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var unmarshaledStatus status.OrderStatus
				err := unmarshaledStatus.UnmarshalJSON([]byte(tt.inputJSON))

				if tt.expectedError != nil {
					if tt.expectedError == domainErrors.ErrInvalidOrderStatus {
						assert.ErrorIs(t, err, domainErrors.ErrInvalidOrderStatus)
					} else {
						assert.Error(t, err)
					}
				} else {
					require.NoError(t, err)
					assert.Equal(t, tt.expected, unmarshaledStatus)
				}
			})
		}
	})
}
