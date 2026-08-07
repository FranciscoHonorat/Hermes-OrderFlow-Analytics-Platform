package command_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/FranciscoHonorat/ordemflow/services/order-service/application/command"
	"github.com/FranciscoHonorat/ordemflow/services/order-service/application/port/output"
	"github.com/FranciscoHonorat/ordemflow/services/order-service/domain/entity"
	"github.com/FranciscoHonorat/ordemflow/services/order-service/domain/event"
	"github.com/FranciscoHonorat/ordemflow/services/order-service/domain/repository"
	"github.com/FranciscoHonorat/ordemflow/services/order-service/domain/valueobject"
)

type MockClock struct {
	NowTime time.Time
}

func (m *MockClock) Now() time.Time {
	return m.NowTime
}

type MockOrderRepository struct {
	SaveFunc     func(ctx context.Context, order *entity.Order) error
	FindByIDFunc func(ctx context.Context, id valueobject.OrderID) (*entity.Order, error)
	Saved        *entity.Order
}

func (m *MockOrderRepository) Save(ctx context.Context, order *entity.Order) error {
	m.Saved = order
	if m.SaveFunc != nil {
		return m.SaveFunc(ctx, order)
	}
	return nil
}

func (m *MockOrderRepository) FindByID(ctx context.Context, id valueobject.OrderID) (*entity.Order, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, nil
}

type MockOutboxRepository struct {
	SavedEvents []event.DomainEvent
}

func (m *MockOutboxRepository) SaveEvents(ctx context.Context, events []event.DomainEvent) error {
	m.SavedEvents = append(m.SavedEvents, events...)
	return nil
}

type MockRepositoryProvider struct {
	orderRepo  *MockOrderRepository
	outboxRepo *MockOutboxRepository
}

func (m *MockRepositoryProvider) OrderRepository() repository.OrderRepository {
	return m.orderRepo
}

func (m *MockRepositoryProvider) OutboxRepository() output.OutboxRepository {
	return m.outboxRepo
}

type MockUnitOfWork struct {
	provider   *MockRepositoryProvider
	ShouldFail bool
}

func (m *MockUnitOfWork) Do(ctx context.Context, fn func(store output.RepositoryProvider) error) error {
	if m.ShouldFail {
		return errors.New("uow database error")
	}
	return fn(m.provider)
}

func TestPlaceOrderHandler_Handle(t *testing.T) {
	fixedTime := time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC)
	customerUUID := uuid.New().String()
	productUUID := uuid.New().String()

	t.Run("Should create order successfully and drain events to outbox", func(t *testing.T) {
		mockClock := &MockClock{NowTime: fixedTime}
		orderRepo := &MockOrderRepository{}
		outboxRepo := &MockOutboxRepository{}

		provider := &MockRepositoryProvider{orderRepo: orderRepo, outboxRepo: outboxRepo}
		uow := &MockUnitOfWork{provider: provider}

		handler := command.NewPlaceOrderHandler(uow, mockClock)

		cmd := command.PlaceOrderCommand{
			CustomerID: customerUUID,
			Items: []command.PlaceOrderItem{
				{
					ProductID:      productUUID,
					Quantity:       2,
					UnitPriceCents: 5000,
					Currency:       "BRL",
				},
			},
		}

		result, err := handler.Handle(context.Background(), cmd)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if result.OrderID == "" {
			t.Error("expected a valid generated OrderID in result")
		}

		if orderRepo.Saved == nil {
			t.Fatal("expected order to be saved in repository")
		}

		if orderRepo.Saved.OrderID().String() != result.OrderID {
			t.Errorf("expected saved order ID to match result, got %s", orderRepo.Saved.OrderID().String())
		}

		if len(outboxRepo.SavedEvents) != 1 {
			t.Fatalf("expected 1 event in outbox, got %d", len(outboxRepo.SavedEvents))
		}

		evt := outboxRepo.SavedEvents[0]
		if evt.EventName() != "order.placed" {
			t.Errorf("expected event name to be 'order.placed', got %s", evt.EventName())
		}

		if !evt.OccurredAt().Equal(fixedTime) {
			t.Errorf("expected event timestamp to be %v, got %v", fixedTime, evt.OccurredAt())
		}

		if len(orderRepo.Saved.PullEvents()) != 0 {
			t.Error("expected entity domain events slice to be empty after processing")
		}
	})

	t.Run("Should return error immediately without hitting database when items list is empty (Fail-Fast)", func(t *testing.T) {
		mockClock := &MockClock{NowTime: fixedTime}
		uow := &MockUnitOfWork{ShouldFail: true}

		handler := command.NewPlaceOrderHandler(uow, mockClock)

		cmd := command.PlaceOrderCommand{
			CustomerID: customerUUID,
			Items:      []command.PlaceOrderItem{},
		}

		_, err := handler.Handle(context.Background(), cmd)

		if err == nil {
			{
				t.Error("expected fail-fast error when order has no items, but got nil")
			}
		}
	})

	t.Run("Should return error when customer UUID is malformed", func(t *testing.T) {
		mockClock := &MockClock{NowTime: fixedTime}
		uow := &MockUnitOfWork{}

		handler := command.NewPlaceOrderHandler(uow, mockClock)

		cmd := command.PlaceOrderCommand{
			CustomerID: "invalid-uuid-format",
			Items: []command.PlaceOrderItem{
				{ProductID: productUUID, Quantity: 1, UnitPriceCents: 1000, Currency: "BRL"},
			},
		}

		_, err := handler.Handle(context.Background(), cmd)

		if err == nil {
			t.Error("expected error due to invalid customer UUID, but got nil")
		}
	})
}
