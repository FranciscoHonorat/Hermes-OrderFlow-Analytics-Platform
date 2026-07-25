package command

import (
	aggregator "github.com/FranciscoHonorat/ordemflow/services/order-service/domain/entity"
)

type AddItemCommand struct {
	Item aggregator.OrderID
}
