package domainErrors

import "errors"

var (
	ErrInvalidQuantity    = errors.New("invalid quantity")
	ErrInsufficientStock  = errors.New("insufficient stock")
	ErrMinimumQuantity    = errors.New("minimum quantity not met")
	ErrInvalidSKU         = errors.New("invalid SKU")
	ErrNoReservedStock    = errors.New("no reserved stock to release")
	ErrCorruptedStockItem = errors.New("stock item is corrupted")
	ErrStockItemNotFound  = errors.New("stock item not found")
)
