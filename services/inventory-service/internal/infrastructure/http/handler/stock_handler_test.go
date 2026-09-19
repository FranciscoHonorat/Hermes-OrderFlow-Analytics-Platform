package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/FranciscoHonorat/ordemflow/services/inventory-service/internal/application/port/input"
	"github.com/FranciscoHonorat/ordemflow/services/inventory-service/internal/infrastructure/http/handler"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type mockReserveUseCase struct {
	err      error
	gotInput input.ReserveStockInput
}

func (m *mockReserveUseCase) Execute(ctx context.Context, in input.ReserveStockInput) error {
	m.gotInput = in
	return m.err
}

type mockReleaseUseCase struct {
	err      error
	gotInput input.ReleaseStockInput
}

func (m *mockReleaseUseCase) Execute(ctx context.Context, in input.ReleaseStockInput) error {
	m.gotInput = in
	return m.err
}

type mockReplenishUseCase struct {
	err      error
	gotInput input.ReplenishStockInput
}

func (m *mockReplenishUseCase) Execute(ctx context.Context, in input.ReplenishStockInput) error {
	m.gotInput = in
	return m.err
}

type mockStockQueries struct {
	dto     *input.StockDTO
	getErr  error
	list    []input.StockDTO
	listErr error
}

func (m *mockStockQueries) GetStockBySKU(ctx context.Context, sku string) (*input.StockDTO, error) {
	return m.dto, m.getErr
}

func (m *mockStockQueries) ListLowStock(ctx context.Context) ([]input.StockDTO, error) {
	return m.list, m.listErr
}

func newJSONContext(t *testing.T, method, path string, body any) (*httptest.ResponseRecorder, *gin.Context) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	var reader *bytes.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(raw)
	} else {
		reader = bytes.NewReader(nil)
	}

	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	return w, c
}

func TestStockHandler_ReserveStock(t *testing.T) {
	t.Run("valid request reserves stock", func(t *testing.T) {
		reserve := &mockReserveUseCase{}
		h := handler.NewStockHandler(&mockStockQueries{}, reserve, &mockReleaseUseCase{}, &mockReplenishUseCase{})

		w, c := newJSONContext(t, "POST", "/stock/reserve", input.ReserveStockInput{SKU: "sku-1", Quantity: 5})
		h.ReserveStock(c)

		require.Equal(t, 200, w.Code)
		require.Equal(t, "sku-1", reserve.gotInput.SKU)
		require.Equal(t, 5, reserve.gotInput.Quantity)
	})

	t.Run("malformed body returns 400", func(t *testing.T) {
		h := handler.NewStockHandler(&mockStockQueries{}, &mockReserveUseCase{}, &mockReleaseUseCase{}, &mockReplenishUseCase{})

		w, c := newJSONContext(t, "POST", "/stock/reserve", nil)
		h.ReserveStock(c)

		require.Equal(t, 400, w.Code)
	})

	t.Run("use case error returns 422", func(t *testing.T) {
		reserve := &mockReserveUseCase{err: errors.New("insufficient stock")}
		h := handler.NewStockHandler(&mockStockQueries{}, reserve, &mockReleaseUseCase{}, &mockReplenishUseCase{})

		w, c := newJSONContext(t, "POST", "/stock/reserve", input.ReserveStockInput{SKU: "sku-1", Quantity: 5})
		h.ReserveStock(c)

		require.Equal(t, 422, w.Code)
	})
}

func TestStockHandler_ReleaseStock(t *testing.T) {
	t.Run("valid request releases stock", func(t *testing.T) {
		release := &mockReleaseUseCase{}
		h := handler.NewStockHandler(&mockStockQueries{}, &mockReserveUseCase{}, release, &mockReplenishUseCase{})

		w, c := newJSONContext(t, "POST", "/stock/release", input.ReleaseStockInput{SKU: "sku-1", Quantity: 3})
		h.ReleaseStock(c)

		require.Equal(t, 200, w.Code)
		require.Equal(t, "sku-1", release.gotInput.SKU)
	})

	t.Run("use case error returns 422", func(t *testing.T) {
		release := &mockReleaseUseCase{err: errors.New("no reserved stock")}
		h := handler.NewStockHandler(&mockStockQueries{}, &mockReserveUseCase{}, release, &mockReplenishUseCase{})

		w, c := newJSONContext(t, "POST", "/stock/release", input.ReleaseStockInput{SKU: "sku-1", Quantity: 3})
		h.ReleaseStock(c)

		require.Equal(t, 422, w.Code)
	})
}

func TestStockHandler_ReplenishStock(t *testing.T) {
	t.Run("valid request replenishes stock", func(t *testing.T) {
		replenish := &mockReplenishUseCase{}
		h := handler.NewStockHandler(&mockStockQueries{}, &mockReserveUseCase{}, &mockReleaseUseCase{}, replenish)

		w, c := newJSONContext(t, "POST", "/stock/replenish", input.ReplenishStockInput{SKU: "sku-1", Quantity: 20})
		h.ReplenishStock(c)

		require.Equal(t, 200, w.Code)
		require.Equal(t, 20, replenish.gotInput.Quantity)
	})
}

func TestStockHandler_GetStockBySKU(t *testing.T) {
	t.Run("found returns 200 with the stock DTO", func(t *testing.T) {
		queries := &mockStockQueries{dto: &input.StockDTO{SKU: "sku-1", Available: 10}}
		h := handler.NewStockHandler(queries, &mockReserveUseCase{}, &mockReleaseUseCase{}, &mockReplenishUseCase{})

		w, c := newJSONContext(t, "GET", "/stock/sku-1", nil)
		c.Params = gin.Params{{Key: "sku", Value: "sku-1"}}
		h.GetStockBySKU(c)

		require.Equal(t, 200, w.Code)

		var got input.StockDTO
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
		require.Equal(t, "sku-1", got.SKU)
		require.Equal(t, 10, got.Available)
	})

	t.Run("not found returns 404", func(t *testing.T) {
		queries := &mockStockQueries{getErr: errors.New("stock item not found")}
		h := handler.NewStockHandler(queries, &mockReserveUseCase{}, &mockReleaseUseCase{}, &mockReplenishUseCase{})

		w, c := newJSONContext(t, "GET", "/stock/missing", nil)
		c.Params = gin.Params{{Key: "sku", Value: "missing"}}
		h.GetStockBySKU(c)

		require.Equal(t, 404, w.Code)
	})
}

func TestStockHandler_ListLowStock(t *testing.T) {
	t.Run("success returns 200 with the list", func(t *testing.T) {
		queries := &mockStockQueries{list: []input.StockDTO{{SKU: "sku-1", Available: 1, MinimumQuantity: 5}}}
		h := handler.NewStockHandler(queries, &mockReserveUseCase{}, &mockReleaseUseCase{}, &mockReplenishUseCase{})

		w, c := newJSONContext(t, "GET", "/stock/low", nil)
		h.ListLowStock(c)

		require.Equal(t, 200, w.Code)

		var got []input.StockDTO
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
		require.Len(t, got, 1)
	})

	t.Run("query error returns 500", func(t *testing.T) {
		queries := &mockStockQueries{listErr: errors.New("db error")}
		h := handler.NewStockHandler(queries, &mockReserveUseCase{}, &mockReleaseUseCase{}, &mockReplenishUseCase{})

		w, c := newJSONContext(t, "GET", "/stock/low", nil)
		h.ListLowStock(c)

		require.Equal(t, 500, w.Code)
	})
}
