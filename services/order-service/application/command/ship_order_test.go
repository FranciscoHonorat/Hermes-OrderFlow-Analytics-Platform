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

func TestShipOrderHandler_Handle(t *testing.T) {
	fixedTime := time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC)
	orderUUID := uuid.New().String()

	t.Run("Should ship order successfully and save to outbox", func(t *testing.T) {
		mockClock := &MockClock{NowTime: fixedTime}
		outboxRepo := &MockOutboxRepository{}

		createTime := fixedTime.Add(-5 * time.Minute)
		confirmTime := fixedTime.Add(-2 * time.Minute)

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
		err = existingOrder.AddItem(item)
		require.NoError(t, err)

		err = existingOrder.Place(createTime)
		require.NoError(t, err)

		_ = existingOrder.PullEvents()

		err = existingOrder.Confirm(confirmTime)
		require.NoError(t, err)

		_ = existingOrder.PullEvents()

		orderRepo := &MockOrderRepository{
			FindByIDFunc: func(ctx context.Context, id valueobject.OrderID) (*entity.Order, error) {
				return existingOrder, nil
			},
		}

		provider := &MockRepositoryProvider{orderRepo: orderRepo, outboxRepo: outboxRepo}
		uow := &MockUnitOfWork{provider: provider}

		hanler := command.NewShipOrderHandler(uow, mockClock)

		cmd := command.ShipOrderCommand{
			OrderID: orderUUID,
		}

		err = hanler.Handle(context.Background(), cmd)
		require.NoError(t, err)

		require.NotNil(t, orderRepo.Saved)

		require.Equal(t, valueobject.OrderStatusShipped, orderRepo.Saved.Status())

		require.Len(t, outboxRepo.SavedEvents, 1)
		evt := outboxRepo.SavedEvents[0]
		require.Equal(t, "order.shipped", evt.EventName())
		require.Equal(t, fixedTime, evt.OccurredAt())
	})
}
