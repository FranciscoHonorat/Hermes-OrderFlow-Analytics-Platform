package command_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/FranciscoHonorat/ordemflow/services/inventory-service/internal/application/command"
	"github.com/FranciscoHonorat/ordemflow/services/inventory-service/internal/application/port/output"
	"github.com/FranciscoHonorat/ordemflow/services/inventory-service/internal/domain/repository"
	"github.com/FranciscoHonorat/ordemflow/services/inventory-service/internal/domain/stock"
	"github.com/FranciscoHonorat/ordemflow/services/inventory-service/internal/domain/valueobject"
	"github.com/FranciscoHonorat/ordemflow/shared/events"
	"github.com/FranciscoHonorat/ordemflow/shared/outbox"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type MockClock struct {
	NowTime time.Time
}

func (m *MockClock) Now() time.Time {
	return m.NowTime
}

type MockStockRepository struct {
	Item  *stock.StockItem
	Saved *stock.StockItem
}

func (m *MockStockRepository) Save(ctx context.Context, item *stock.StockItem) error {
	m.Saved = item
	return nil
}

func (m *MockStockRepository) FindBySKU(ctx context.Context, sku valueobject.SKU) (*stock.StockItem, error) {
	if m.Item == nil {
		return nil, errors.New("stock item not found")
	}
	return m.Item, nil
}

type MockOutboxRepository struct {
	SavedEvents []events.DomainEvent
}

func (m *MockOutboxRepository) SaveEvents(ctx context.Context, evts []events.DomainEvent) error {
	m.SavedEvents = append(m.SavedEvents, evts...)
	return nil
}

func (m *MockOutboxRepository) FetchUnprocessed(ctx context.Context, limit int) ([]outbox.Row, error) {
	return nil, nil
}

func (m *MockOutboxRepository) MarkProcessed(ctx context.Context, ids []uuid.UUID) error {
	return nil
}

type MockRepositoryProvider struct {
	stockRepo  *MockStockRepository
	outboxRepo *MockOutboxRepository
}

func (m *MockRepositoryProvider) StockRepository() repository.StockRepository {
	return m.stockRepo
}

func (m *MockRepositoryProvider) OutboxRepository() output.OutboxRepository {
	return m.outboxRepo
}

type MockUnitOfWork struct {
	provider *MockRepositoryProvider
}

func (m *MockUnitOfWork) Do(ctx context.Context, fn func(store output.RepositoryProvider) error) error {
	return fn(m.provider)
}

func newTestItem(t *testing.T, available, minimumQuantity int, now time.Time) *stock.StockItem {
	t.Helper()
	sku, err := valueobject.NewSKU("item1")
	require.NoError(t, err)

	item, err := stock.NewStockItem(sku, available, minimumQuantity, now)
	require.NoError(t, err)
	return item
}

func TestReserveStockHandler_Handle(t *testing.T) {
	fixedTime := time.Date(2026, 9, 18, 12, 0, 0, 0, time.UTC)

	t.Run("reserves stock and drains events to outbox", func(t *testing.T) {
		stockRepo := &MockStockRepository{Item: newTestItem(t, 10, 2, fixedTime)}
		outboxRepo := &MockOutboxRepository{}
		uow := &MockUnitOfWork{provider: &MockRepositoryProvider{stockRepo: stockRepo, outboxRepo: outboxRepo}}

		handler := command.NewReserveStockHandler(uow, &MockClock{NowTime: fixedTime})

		err := handler.Handle(context.Background(), command.ReserveStockCommand{SKU: "item1", Quantity: 5})
		require.NoError(t, err)

		require.NotNil(t, stockRepo.Saved)
		require.Equal(t, 5, stockRepo.Saved.Available())
		require.Equal(t, 5, stockRepo.Saved.Reserved())

		require.Len(t, outboxRepo.SavedEvents, 1)
		require.Equal(t, "stock.reserved", outboxRepo.SavedEvents[0].EventName())
	})

	t.Run("returns error without saving when stock is insufficient", func(t *testing.T) {
		stockRepo := &MockStockRepository{Item: newTestItem(t, 3, 2, fixedTime)}
		outboxRepo := &MockOutboxRepository{}
		uow := &MockUnitOfWork{provider: &MockRepositoryProvider{stockRepo: stockRepo, outboxRepo: outboxRepo}}

		handler := command.NewReserveStockHandler(uow, &MockClock{NowTime: fixedTime})

		err := handler.Handle(context.Background(), command.ReserveStockCommand{SKU: "item1", Quantity: 5})
		require.Error(t, err)
		require.Nil(t, stockRepo.Saved)
		require.Empty(t, outboxRepo.SavedEvents)
	})

	t.Run("returns error for invalid SKU without touching the repository", func(t *testing.T) {
		stockRepo := &MockStockRepository{Item: newTestItem(t, 10, 2, fixedTime)}
		outboxRepo := &MockOutboxRepository{}
		uow := &MockUnitOfWork{provider: &MockRepositoryProvider{stockRepo: stockRepo, outboxRepo: outboxRepo}}

		handler := command.NewReserveStockHandler(uow, &MockClock{NowTime: fixedTime})

		err := handler.Handle(context.Background(), command.ReserveStockCommand{SKU: "   ", Quantity: 5})
		require.Error(t, err)
		require.Nil(t, stockRepo.Saved)
	})
}

func TestReleaseStockHandler_Handle(t *testing.T) {
	fixedTime := time.Date(2026, 9, 18, 12, 0, 0, 0, time.UTC)

	t.Run("releases previously reserved stock", func(t *testing.T) {
		item := newTestItem(t, 10, 2, fixedTime)
		require.NoError(t, item.Reserve(5, fixedTime))
		item.ClearEvents()

		stockRepo := &MockStockRepository{Item: item}
		outboxRepo := &MockOutboxRepository{}
		uow := &MockUnitOfWork{provider: &MockRepositoryProvider{stockRepo: stockRepo, outboxRepo: outboxRepo}}

		handler := command.NewReleaseStockHandler(uow, &MockClock{NowTime: fixedTime})

		err := handler.Handle(context.Background(), command.ReleaseStockCommand{SKU: "item1", Quantity: 3})
		require.NoError(t, err)

		require.Equal(t, 8, stockRepo.Saved.Available())
		require.Equal(t, 2, stockRepo.Saved.Reserved())
		require.Len(t, outboxRepo.SavedEvents, 1)
		require.Equal(t, "stock.released", outboxRepo.SavedEvents[0].EventName())
	})
}

func TestReplenishStockHandler_Handle(t *testing.T) {
	fixedTime := time.Date(2026, 9, 18, 12, 0, 0, 0, time.UTC)

	t.Run("increases available stock", func(t *testing.T) {
		stockRepo := &MockStockRepository{Item: newTestItem(t, 10, 2, fixedTime)}
		outboxRepo := &MockOutboxRepository{}
		uow := &MockUnitOfWork{provider: &MockRepositoryProvider{stockRepo: stockRepo, outboxRepo: outboxRepo}}

		handler := command.NewReplenishStockHandler(uow, &MockClock{NowTime: fixedTime})

		err := handler.Handle(context.Background(), command.ReplenishStockCommand{SKU: "item1", Quantity: 20})
		require.NoError(t, err)

		require.Equal(t, 30, stockRepo.Saved.Available())
		require.Len(t, outboxRepo.SavedEvents, 1)
		require.Equal(t, "stock.replenished", outboxRepo.SavedEvents[0].EventName())
	})
}
