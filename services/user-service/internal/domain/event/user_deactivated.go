package event

import (
	"time"

	"github.com/FranciscoHonorat/ordemflow/shared/events"
)

var _ events.DomainEvent = (*UserDeactivated)(nil)

type UserDeactivated struct {
	events.BaseEvent
}

func NewUserDeactivated(userID string, occurredAt time.Time) UserDeactivated {
	return UserDeactivated{
		BaseEvent: events.NewBaseEvent("user.deactivated", userID, occurredAt),
	}
}
