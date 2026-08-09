package command_test

import (
	"context"
	"testing"
	"time"

	"github.com/FranciscoHonorat/ordemflow/services/order-service/application/command"
	"github.com/FranciscoHonorat/ordemflow/services/order-service/domain/entity"
	"github.com/FranciscoHonorat/ordemflow/services/order-service/domain/valueobject"
	"github.com/google/uuid"
)

func TestCancelOrderHandler_Handle(t *testing.T) {
	fixedTime := time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC)
	orderUUID := uuid.New().String()

	t.Run("Should cancel order successfully and save to outbox", func(t *testing.T) {
		mockClock := &MockClock{NowTime: fixedTime}
		outboxRepo := &MockOutboxRepository{}

		createTime := fixedTime.Add(-5 * time.Minute)

		orderID, err := valueobject.NewOrderID(uuid.MustParse(orderUUID))
		if err != nil {
			t.Fatalf("failed to create order ID: %v", err)
		}
		customerID, err := valueobject.NewCustomerID(uuid.New())
		if err != nil {
			t.Fatalf("failed to create customer ID: %v", err)
		}
		existingOrder, err := entity.NewOrder(orderID, customerID)
		if err != nil {
			t.Fatalf("failed to create order: %v", err)
		}

		productID, err := valueobject.NewProductID(uuid.New())
		if err != nil {
			t.Fatalf("failed to create product ID: %v", err)
		}
		price, err := valueobject.NewMoney(1000, "USD")
		if err != nil {
			t.Fatalf("failed to create price: %v", err)
		}
		quantity, err := valueobject.NewQuantity(2)
		if err != nil {
			t.Fatalf("failed to create quantity: %v", err)
		}
		item, err := valueobject.NewOrderItem(productID, price, quantity)
		if err != nil {
			t.Fatalf("failed to create order item: %v", err)
		}
		if err := existingOrder.AddItem(item); err != nil {
			t.Fatalf("failed to add item to order: %v", err)
		}

		if err := existingOrder.Place(createTime); err != nil {
			t.Fatalf("failed to place order: %v", err)
		}
		_ = existingOrder.PullEvents()

		orderRepo := &MockOrderRepository{
			FindByIDFunc: func(ctx context.Context, id valueobject.OrderID) (*entity.Order, error) {
				return existingOrder, nil
			},
		}

		provider := &MockRepositoryProvider{orderRepo: orderRepo, outboxRepo: outboxRepo}
		uow := &MockUnitOfWork{provider: provider}

		handler := command.NewCancelOrderHandler(uow, mockClock)

		cmd := command.CancelOrderCommand{
			OrderID: orderUUID,
		}

		err = handler.Handle(context.Background(), cmd)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(outboxRepo.SavedEvents) != 1 {
			t.Errorf("expected 1 event in outbox, got %d", len(outboxRepo.SavedEvents))
		}

		evt := outboxRepo.SavedEvents[0]
		if evt.EventName() != "order.cancelled" {
			t.Errorf("expected event type 'order.cancelled', got '%s'", evt.EventName())
		}
	})
}
