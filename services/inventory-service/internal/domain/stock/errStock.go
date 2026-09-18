package stock

import "errors"

var (
	ErrInvalidQuantity   = errors.New("invalid quantity")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrMinimumQuantity   = errors.New("minimum quantity not met")
)
