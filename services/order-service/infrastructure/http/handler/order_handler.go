package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/FranciscoHonorat/ordemflow/services/order-service/application/port/input"
	"github.com/google/uuid"
)

type OrderIDRequest struct {
	OrderID string `json:"order_id"`
}

type OrderHandler struct {
	orderQueries        input.OrderQueries
	placeOrderUseCase   input.PlaceOrderUseCase
	confirmOrderUseCase input.ConfirmOrderUseCase
	cancelOrderUseCase  input.CancelOrderUseCase
	shipOrderUseCase    input.ShipOrderUseCase
}

func NewOrderHandler(
	queries input.OrderQueries,
	place input.PlaceOrderUseCase,
	confirm input.ConfirmOrderUseCase,
	cancel input.CancelOrderUseCase,
	ship input.ShipOrderUseCase,
) *OrderHandler {
	return &OrderHandler{
		orderQueries:        queries,
		placeOrderUseCase:   place,
		confirmOrderUseCase: confirm,
		cancelOrderUseCase:  cancel,
		shipOrderUseCase:    ship,
	}
}

func (h *OrderHandler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Faz o decode direto na struct de entrada da aplicação
	var req input.CreateOrderInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Malformed JSON body")
		return
	}

	// Invoca a interface da aplicação
	resultID, err := h.placeOrderUseCase.Execute(r.Context(), req)
	if err != nil {
		h.respondWithError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	h.respondWithJSON(w, http.StatusCreated, map[string]string{"order_id": resultID.String()})
}

func (h *OrderHandler) ConfirmOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req OrderIDRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Malformed JSON body")
		return
	}

	orderUUID, err := uuid.Parse(req.OrderID)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid order ID format")
		return
	}

	if err := h.confirmOrderUseCase.Execute(r.Context(), orderUUID); err != nil {
		h.respondWithError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "Order confirmed successfully"})
}

func (h *OrderHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req OrderIDRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Malformed JSON body")
		return
	}

	orderUUID, err := uuid.Parse(req.OrderID)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid order ID format")
		return
	}

	if err := h.cancelOrderUseCase.Execute(r.Context(), orderUUID); err != nil {
		h.respondWithError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "Order cancelled successfully"})
}

func (h *OrderHandler) ShipOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req OrderIDRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Malformed JSON body")
		return
	}

	orderUUID, err := uuid.Parse(req.OrderID)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid order ID format")
		return
	}

	if err := h.shipOrderUseCase.Execute(r.Context(), orderUUID); err != nil {
		h.respondWithError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "Order shipped successfully"})
}

func (h *OrderHandler) GetOrderByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	idStr := r.URL.Query().Get("id")
	orderID, err := uuid.Parse(idStr)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid order ID format")
		return
	}

	orderDTO, err := h.orderQueries.GetOrderByID(r.Context(), orderID)
	if err != nil {
		h.respondWithError(w, http.StatusNotFound, "Order not found")
		return
	}

	h.respondWithJSON(w, http.StatusOK, orderDTO)
}

func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit <= 0 {
		limit = 10
	}
	offset, _ := strconv.Atoi(q.Get("offset"))

	var customerID *uuid.UUID
	if cidStr := q.Get("customer_id"); cidStr != "" {
		if cid, err := uuid.Parse(cidStr); err == nil {
			customerID = &cid
		} else {
			h.respondWithError(w, http.StatusBadRequest, "Invalid customer ID format")
			return
		}
	}

	orders, err := h.orderQueries.ListOrders(r.Context(), customerID, limit, offset)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to list orders")
		return
	}

	h.respondWithJSON(w, http.StatusOK, orders)
}

func (h *OrderHandler) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, map[string]string{"error": message})
}

func (h *OrderHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}
