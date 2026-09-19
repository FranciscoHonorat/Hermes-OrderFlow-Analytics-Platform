package event

import (
	"time"

	"github.com/FranciscoHonorat/ordemflow/shared/events"
)

var _ events.DomainEvent = (*UserRoleChanged)(nil)

type UserRoleChanged struct {
	events.BaseEvent
	OldRole string
	NewRole string
}

func NewUserRoleChanged(userID, oldRole, newRole string, occurredAt time.Time) UserRoleChanged {
	return UserRoleChanged{
		BaseEvent: events.NewBaseEvent("user.role_changed", userID, occurredAt),
		OldRole:   oldRole,
		NewRole:   newRole,
	}
}
