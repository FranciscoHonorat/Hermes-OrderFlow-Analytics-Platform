package query_test

import (
	"context"
	"testing"
	"time"

	"github.com/FranciscoHonorat/ordemflow/services/order-service/application/query"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockOrderQueries struct {
	mock.Mock
}

func (m *MockOrderQueries) GetOrderByID(ctx context.Context, orderID uuid.UUID) (*query.OrderDTO, error) {
	args := m.Called(ctx, orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*query.OrderDTO), args.Error(1)
}

func (m *MockOrderQueries) ListOrders(ctx context.Context, customerID *uuid.UUID, limit, offset int) ([]query.OrderDTO, error) {
	args := m.Called(ctx, customerID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]query.OrderDTO), args.Error(1)
}

func TestOrderQueries(t *testing.T) {
	ctx := context.Background()
	orderID := uuid.New()

	expectedOrder := &query.OrderDTO{
		ID:         orderID,
		CustomerID: uuid.New(),
		Status:     "PENDING",
		Total:      1000,
		Currency:   "USD",
		Items: []query.ItemDTO{
			{
				ProductID: uuid.New(),
				Quantity:  2,
				Price:     500,
			},
		},
		CreatedAt: time.Now(),
	}

	t.Run("GetOrderByID returns expected order", func(t *testing.T) {
		mockQueries := new(MockOrderQueries)
		mockQueries.On("GetOrderByID", ctx, orderID).Return(expectedOrder, nil)

		res, err := mockQueries.GetOrderByID(ctx, orderID)

		assert.NoError(t, err)
		assert.Equal(t, expectedOrder, res)
		mockQueries.AssertExpectations(t)
	})

	t.Run("ListOrders returns expected orders", func(t *testing.T) {
		mockQueries := new(MockOrderQueries)
		mockQueries.On("ListOrders", ctx, (*uuid.UUID)(nil), 10, 0).Return([]query.OrderDTO{*expectedOrder}, nil)

		res, err := mockQueries.ListOrders(ctx, nil, 10, 0)

		assert.NoError(t, err)
		assert.Len(t, res, 1)
		assert.Equal(t, expectedOrder.ID, res[0].ID)
		mockQueries.AssertExpectations(t)
	})

	t.Run("GetOrderByID returns error", func(t *testing.T) {
		mockQueries := new(MockOrderQueries)
		mockQueries.On("GetOrderByID", ctx, orderID).Return((*query.OrderDTO)(nil), assert.AnError)

		res, err := mockQueries.GetOrderByID(ctx, orderID)

		assert.ErrorIs(t, err, assert.AnError)
		assert.Nil(t, res)
		mockQueries.AssertExpectations(t)
	})

	t.Run("ListOrders returns error", func(t *testing.T) {
		mockQueries := new(MockOrderQueries)
		mockQueries.On("ListOrders", ctx, (*uuid.UUID)(nil), 10, 0).Return(([]query.OrderDTO)(nil), assert.AnError)

		res, err := mockQueries.ListOrders(ctx, nil, 10, 0)

		assert.ErrorIs(t, err, assert.AnError)
		assert.Nil(t, res)
		mockQueries.AssertExpectations(t)
	})
}
