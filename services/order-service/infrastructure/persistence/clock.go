package persistence

import (
	"time"

	"github.com/FranciscoHonorat/ordemflow/services/order-service/application/port/output"
)

var _ output.Clock = (*RealClock)(nil)

type RealClock struct{}

func NewRealClock() output.Clock {
	return &RealClock{}
}

func (c *RealClock) Now() time.Time {
	return time.Now()
}
