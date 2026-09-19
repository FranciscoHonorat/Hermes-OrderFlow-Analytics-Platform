package event

import (
	"time"

	"github.com/FranciscoHonorat/ordemflow/shared/events"
)

var _ events.DomainEvent = (*PasswordChanged)(nil)

type PasswordChanged struct {
	events.BaseEvent
}

func NewPasswordChanged(userID string, occurredAt time.Time) PasswordChanged {
	return PasswordChanged{
		BaseEvent: events.NewBaseEvent("user.password_changed", userID, occurredAt),
	}
}
