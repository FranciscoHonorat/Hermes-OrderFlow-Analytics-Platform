package event

import (
	"time"

	"github.com/FranciscoHonorat/ordemflow/shared/events"
)

var _ events.DomainEvent = (*UserActivated)(nil)

type UserActivated struct {
	events.BaseEvent
}

func NewUserActivated(userID string, occurredAt time.Time) UserActivated {
	return UserActivated{
		BaseEvent: events.NewBaseEvent("user.activated", userID, occurredAt),
	}
}
