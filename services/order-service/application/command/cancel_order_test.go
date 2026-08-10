package command_test

import (
	"context"
	"testing"
	"time"

	"github.com/FranciscoHonorat/ordemflow/services/order-service/application/command"
	"github.com/FranciscoHonorat/ordemflow/services/order-service/domain/entity"
	"github.com/FranciscoHonorat/ordemflow/services/order-service/domain/valueobject"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCancelOrderHandler_Handle(t *testing.T) {
	fixedTime := time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC)
	orderUUID := uuid.New().String()

	t.Run("Should cancel order successfully and save to outbox", func(t *testing.T) {
		mockClock := &MockClock{NowTime: fixedTime}
		outboxRepo := &MockOutboxRepository{}

		createTime := fixedTime.Add(-5 * time.Minute)

		orderID, err := valueobject.NewOrderID(uuid.MustParse(orderUUID))
		require.NoError(t, err)

		customerID, err := valueobject.NewCustomerID(uuid.New())
		require.NoError(t, err)

		existingOrder, err := entity.NewOrder(orderID, customerID)
		require.NoError(t, err)

		productID, err := valueobject.NewProductID(uuid.New())
		require.NoError(t, err)

		price, err := valueobject.NewMoney(1000, "USD")
		require.NoError(t, err)

		quantity, err := valueobject.NewQuantity(2)
		require.NoError(t, err)

		item, err := valueobject.NewOrderItem(productID, price, quantity)
		require.NoError(t, err)

		require.NoError(t, existingOrder.AddItem(item))

		require.NoError(t, existingOrder.Place(createTime))
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

		require.NoError(t, err)

		require.Equal(t, valueobject.OrderStatusCancelled, existingOrder.Status())

		require.Len(t, outboxRepo.SavedEvents, 1)

		evt := outboxRepo.SavedEvents[0]
		require.Equal(t, "order.cancelled", evt.EventName())
	})
}
