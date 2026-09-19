package stock

import (
	"time"

	domainErrors "github.com/FranciscoHonorat/ordemflow/services/inventory-service/internal/domain/domain-errors"
	"github.com/FranciscoHonorat/ordemflow/services/inventory-service/internal/domain/event"
	"github.com/FranciscoHonorat/ordemflow/services/inventory-service/internal/domain/valueobject"
	"github.com/FranciscoHonorat/ordemflow/shared/events"
)

// StockItem é o agregado raiz do bounded context de inventário. Toda
// mutação de estado acontece através dos seus métodos, nunca por
// atribuição direta aos campos (ver ARCHITECTURE.md, seção 5).
type StockItem struct {
	sku             valueobject.SKU
	available       int
	reserved        int
	minimumQuantity int
	updatedAt       time.Time
	domainEvents    []events.DomainEvent
}

func NewStockItem(sku valueobject.SKU, available, minimumQuantity int, now time.Time) (*StockItem, error) {
	if sku.IsZero() {
		return nil, domainErrors.ErrInvalidSKU
	}
	if available < 0 {
		return nil, domainErrors.ErrInvalidQuantity
	}
	if minimumQuantity < 0 {
		return nil, domainErrors.ErrMinimumQuantity
	}

	return &StockItem{
		sku:             sku,
		available:       available,
		minimumQuantity: minimumQuantity,
		updatedAt:       now.UTC(),
	}, nil
}

// RestoreStockItem reconstrói um StockItem a partir de dados já persistidos,
// sem revalidar invariantes de criação nem emitir eventos.
func RestoreStockItem(sku valueobject.SKU, available, reserved, minimumQuantity int, updatedAt time.Time) *StockItem {
	return &StockItem{
		sku:             sku,
		available:       available,
		reserved:        reserved,
		minimumQuantity: minimumQuantity,
		updatedAt:       updatedAt,
	}
}

func (s *StockItem) Reserve(quantity int, now time.Time) error {
	if !s.isValid() {
		return domainErrors.ErrCorruptedStockItem
	}
	if quantity <= 0 {
		return domainErrors.ErrInvalidQuantity
	}
	if quantity > s.available {
		return domainErrors.ErrInsufficientStock
	}

	s.available -= quantity
	s.reserved += quantity
	s.updatedAt = now.UTC()

	s.addEvent(event.NewStockReserved(s.sku.String(), quantity, s.available, now))

	if s.available < s.minimumQuantity {
		s.addEvent(event.NewLowStockDetected(s.sku.String(), s.available, s.minimumQuantity, now))
	}

	return nil
}

func (s *StockItem) Release(quantity int, now time.Time) error {
	if !s.isValid() {
		return domainErrors.ErrCorruptedStockItem
	}
	if quantity <= 0 {
		return domainErrors.ErrInvalidQuantity
	}
	if quantity > s.reserved {
		return domainErrors.ErrNoReservedStock
	}

	s.reserved -= quantity
	s.available += quantity
	s.updatedAt = now.UTC()

	s.addEvent(event.NewStockReleased(s.sku.String(), quantity, s.available, now))

	return nil
}

func (s *StockItem) Replenish(quantity int, now time.Time) error {
	if !s.isValid() {
		return domainErrors.ErrCorruptedStockItem
	}
	if quantity <= 0 {
		return domainErrors.ErrInvalidQuantity
	}

	s.available += quantity
	s.updatedAt = now.UTC()

	s.addEvent(event.NewStockReplenished(s.sku.String(), quantity, s.available, now))

	return nil
}

func (s *StockItem) isValid() bool {
	return !s.sku.IsZero() && s.available >= 0 && s.reserved >= 0
}

func (s *StockItem) addEvent(evt events.DomainEvent) {
	s.domainEvents = append(s.domainEvents, evt)
}

func (s *StockItem) DomainEvents() []events.DomainEvent {
	cp := make([]events.DomainEvent, len(s.domainEvents))
	copy(cp, s.domainEvents)
	return cp
}

func (s *StockItem) PullEvents() []events.DomainEvent {
	evts := s.DomainEvents()
	s.ClearEvents()
	return evts
}

func (s *StockItem) ClearEvents() {
	s.domainEvents = nil
}

func (s *StockItem) SKU() valueobject.SKU {
	return s.sku
}

func (s *StockItem) Available() int {
	return s.available
}

func (s *StockItem) Reserved() int {
	return s.reserved
}

func (s *StockItem) MinimumQuantity() int {
	return s.minimumQuantity
}

func (s *StockItem) UpdatedAt() time.Time {
	return s.updatedAt
}
