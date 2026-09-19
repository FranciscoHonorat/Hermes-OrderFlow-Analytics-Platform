package event

import (
	"time"

	"github.com/FranciscoHonorat/ordemflow/shared/events"
)

var _ events.DomainEvent = (*UserRegistered)(nil)

type UserRegistered struct {
	events.BaseEvent
	Email string
	Role  string
}

func NewUserRegistered(userID, email, role string, occurredAt time.Time) UserRegistered {
	return UserRegistered{
		BaseEvent: events.NewBaseEvent("user.registered", userID, occurredAt),
		Email:     email,
		Role:      role,
	}
}
