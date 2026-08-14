package event_test

import (
	"testing"
	"time"

	"github.com/FranciscoHonorat/ordemflow/services/order-service/domain/event"
	"github.com/stretchr/testify/assert"
)

func TestBaseEvent(t *testing.T) {
	t.Run("should create a new BaseEvent with the correct values", func(t *testing.T) {
		tests := []struct {
			name        string
			eventName   string
			aggregateId string
			occurredAt  time.Time
			expectError bool
		}{
			{
				name:        "Valid BaseEvent",
				eventName:   "OrderCreated",
				aggregateId: "12345",
				occurredAt:  time.Now().UTC(),
				expectError: false,
			},
			{
				name:        "Invalid BaseEvent with empty event name",
				eventName:   "",
				aggregateId: "12345",
				occurredAt:  time.Now().UTC(),
				expectError: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				baseEvent := event.NewBaseEvent(tt.eventName, tt.aggregateId, tt.occurredAt)

				if tt.expectError {
					assert.Error(t, baseEvent.Validate())
				} else {
					assert.NoError(t, baseEvent.Validate())
					assert.Equal(t, tt.eventName, baseEvent.EventName())
				}

			})
		}
	})
}
