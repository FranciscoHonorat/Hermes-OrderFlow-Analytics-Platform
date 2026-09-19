package event_test

import (
	"testing"
	"time"

	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/domain/event"
	"github.com/stretchr/testify/assert"
)

var fixedTime = time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)

func TestUserRegistered(t *testing.T) {
	t.Run("Test NewUserRegistered method", func(t *testing.T) {
		evt := event.NewUserRegistered("user-1", "user@example.com", "user", fixedTime)

		assert.Equal(t, "user.registered", evt.EventName())
		assert.Equal(t, "user-1", evt.AggregateId())
		assert.Equal(t, fixedTime, evt.OccurredAt())
		assert.Equal(t, "user@example.com", evt.Email)
		assert.Equal(t, "user", evt.Role)
	})
}

func TestUserRoleChanged(t *testing.T) {
	t.Run("Test NewUserRoleChanged method", func(t *testing.T) {
		evt := event.NewUserRoleChanged("user-1", "user", "admin", fixedTime)

		assert.Equal(t, "user.role_changed", evt.EventName())
		assert.Equal(t, "user", evt.OldRole)
		assert.Equal(t, "admin", evt.NewRole)
	})
}

func TestUserDeactivated(t *testing.T) {
	t.Run("Test NewUserDeactivated method", func(t *testing.T) {
		evt := event.NewUserDeactivated("user-1", fixedTime)

		assert.Equal(t, "user.deactivated", evt.EventName())
		assert.Equal(t, "user-1", evt.AggregateId())
	})
}

func TestUserActivated(t *testing.T) {
	t.Run("Test NewUserActivated method", func(t *testing.T) {
		evt := event.NewUserActivated("user-1", fixedTime)

		assert.Equal(t, "user.activated", evt.EventName())
		assert.Equal(t, "user-1", evt.AggregateId())
	})
}

func TestPasswordChanged(t *testing.T) {
	t.Run("Test NewPasswordChanged method", func(t *testing.T) {
		evt := event.NewPasswordChanged("user-1", fixedTime)

		assert.Equal(t, "user.password_changed", evt.EventName())
		assert.Equal(t, "user-1", evt.AggregateId())
	})
}
