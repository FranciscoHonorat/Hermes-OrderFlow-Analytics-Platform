package stock

type StockItem struct {
	SKU             string
	Available       int
	Reserved        int
	MinimumQuantity int
}

func (s *StockItem) Reserve(quantity int) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}

	if quantity > s.Available {
		return ErrInsufficientStock
	}

	s.Available -= quantity
	s.Reserved += quantity

	return nil
}
