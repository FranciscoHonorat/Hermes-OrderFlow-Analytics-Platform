package relay_test

import (
	"context"
	"errors"
	"testing"

	"github.com/FranciscoHonorat/ordemflow/services/cdc-connector/internal/relay"
	"github.com/FranciscoHonorat/ordemflow/shared/outbox"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeOutboxReader struct {
	rows      []outbox.Row
	fetchErr  error
	markErr   error
	markedIDs []uuid.UUID
}

func (f *fakeOutboxReader) FetchUnprocessed(ctx context.Context, limit int) ([]outbox.Row, error) {
	if f.fetchErr != nil {
		return nil, f.fetchErr
	}
	return f.rows, nil
}

func (f *fakeOutboxReader) MarkProcessed(ctx context.Context, ids []uuid.UUID) error {
	f.markedIDs = append(f.markedIDs, ids...)
	return f.markErr
}

type publishedMessage struct {
	topic string
	key   []byte
	value []byte
}

type fakePublisher struct {
	failTopics map[string]bool
	published  []publishedMessage
}

func (f *fakePublisher) Publish(ctx context.Context, topic string, key, value []byte) error {
	if f.failTopics[topic] {
		return errors.New("broker unavailable")
	}
	f.published = append(f.published, publishedMessage{topic: topic, key: key, value: value})
	return nil
}

func newRow(id uuid.UUID, aggregateID, eventType string) outbox.Row {
	return outbox.Row{
		ID:          id,
		AggregateID: aggregateID,
		Type:        eventType,
		Payload:     []byte(`{"foo":"bar"}`),
	}
}

func TestRelay(t *testing.T) {
	t.Run("Test RunOnce", func(t *testing.T) {
		reservedID := uuid.New()
		reservedRow := newRow(reservedID, "sku-1", "stock.reserved")

		placedID := uuid.New()
		placedRow := newRow(placedID, "order-1", "order.placed")

		tests := []struct {
			name              string
			source            fakeOutboxReader
			failTopics        map[string]bool
			expectedPublished int
			expectMarked      []uuid.UUID
			expectError       bool
		}{
			{
				name:              "publishes unprocessed rows and marks them processed",
				source:            fakeOutboxReader{rows: []outbox.Row{reservedRow}},
				expectedPublished: 1,
				expectMarked:      []uuid.UUID{reservedID},
			},
			{
				name:              "no unprocessed rows: nothing published, nothing marked",
				source:            fakeOutboxReader{rows: nil},
				expectedPublished: 0,
				expectMarked:      nil,
			},
			{
				name:              "publish failure: row is not marked processed",
				source:            fakeOutboxReader{rows: []outbox.Row{reservedRow}},
				failTopics:        map[string]bool{"inventory-service.stock.reserved": true},
				expectedPublished: 0,
				expectMarked:      nil,
				expectError:       true,
			},
			{
				name:              "one row fails, the other still gets published and marked",
				source:            fakeOutboxReader{rows: []outbox.Row{reservedRow, placedRow}},
				failTopics:        map[string]bool{"inventory-service.stock.reserved": true},
				expectedPublished: 1,
				expectMarked:      []uuid.UUID{placedID},
				expectError:       true,
			},
			{
				name:              "fetch failure: nothing published, error surfaced",
				source:            fakeOutboxReader{fetchErr: errors.New("connection refused")},
				expectedPublished: 0,
				expectMarked:      nil,
				expectError:       true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				source := tt.source
				publisher := &fakePublisher{failTopics: tt.failTopics}

				r := relay.New(publisher, 10, relay.Source{Name: "inventory-service", Outbox: &source})

				published, err := r.RunOnce(context.Background())

				assert.Equal(t, tt.expectedPublished, published)
				assert.ElementsMatch(t, tt.expectMarked, source.markedIDs)

				if tt.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})

	t.Run("Test RunOnce with multiple sources", func(t *testing.T) {
		orderSource := &fakeOutboxReader{rows: []outbox.Row{newRow(uuid.New(), "order-1", "order.placed")}}
		inventorySource := &fakeOutboxReader{fetchErr: errors.New("connection refused")}
		publisher := &fakePublisher{}

		r := relay.New(publisher, 10,
			relay.Source{Name: "order-service", Outbox: orderSource},
			relay.Source{Name: "inventory-service", Outbox: inventorySource},
		)

		published, err := r.RunOnce(context.Background())

		require.Error(t, err, "a failure in one source must still surface")
		assert.Equal(t, 1, published, "a failure in one source must not stop the others from being processed")
		assert.Len(t, orderSource.markedIDs, 1)
		assert.Equal(t, "order-service.order.placed", publisher.published[0].topic)
	})
}
