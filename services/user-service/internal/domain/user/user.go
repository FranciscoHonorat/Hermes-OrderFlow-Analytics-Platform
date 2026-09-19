package user

import (
	"time"

	domainErrors "github.com/FranciscoHonorat/ordemflow/services/user-service/internal/domain/domain-errors"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/domain/event"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/domain/valueobject"
	"github.com/FranciscoHonorat/ordemflow/shared/events"
)

// User é o agregado raiz do bounded context de identidade. Toda mutação
// de estado acontece através dos seus métodos, nunca por atribuição
// direta aos campos (ver ARCHITECTURE.md, seção 5).
type User struct {
	id           valueobject.UserID
	email        valueobject.Email
	passwordHash string
	role         valueobject.Role
	active       bool
	createdAt    time.Time
	updatedAt    time.Time
	domainEvents []events.DomainEvent
}

func NewUser(id valueobject.UserID, email valueobject.Email, passwordHash string, role valueobject.Role, now time.Time) (*User, error) {
	if id.IsZero() {
		return nil, domainErrors.ErrInvalidUserID
	}
	if email.IsZero() {
		return nil, domainErrors.ErrInvalidEmail
	}
	if passwordHash == "" {
		return nil, domainErrors.ErrInvalidPasswordHash
	}
	if role.IsZero() {
		return nil, domainErrors.ErrInvalidRole
	}

	u := &User{
		id:           id,
		email:        email,
		passwordHash: passwordHash,
		role:         role,
		active:       true,
		createdAt:    now.UTC(),
		updatedAt:    now.UTC(),
	}

	u.addEvent(event.NewUserRegistered(id.String(), email.String(), role.String(), now))

	return u, nil
}

// RestoreUser reconstrói um User a partir de dados já persistidos, sem
// revalidar invariantes de criação nem emitir eventos.
func RestoreUser(id valueobject.UserID, email valueobject.Email, passwordHash string, role valueobject.Role, active bool, createdAt, updatedAt time.Time) *User {
	return &User{
		id:           id,
		email:        email,
		passwordHash: passwordHash,
		role:         role,
		active:       active,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}
}

func (u *User) ChangeRole(newRole valueobject.Role, now time.Time) error {
	if !u.isValid() {
		return domainErrors.ErrCorruptedUser
	}
	if newRole.IsZero() {
		return domainErrors.ErrInvalidRole
	}
	if newRole.Equal(u.role) {
		return domainErrors.ErrRoleUnchanged
	}

	oldRole := u.role
	u.role = newRole
	u.updatedAt = now.UTC()

	u.addEvent(event.NewUserRoleChanged(u.id.String(), oldRole.String(), newRole.String(), now))

	return nil
}

func (u *User) Deactivate(now time.Time) error {
	if !u.isValid() {
		return domainErrors.ErrCorruptedUser
	}
	if !u.active {
		return domainErrors.ErrUserAlreadyInactive
	}

	u.active = false
	u.updatedAt = now.UTC()

	u.addEvent(event.NewUserDeactivated(u.id.String(), now))

	return nil
}

func (u *User) Activate(now time.Time) error {
	if !u.isValid() {
		return domainErrors.ErrCorruptedUser
	}
	if u.active {
		return domainErrors.ErrUserAlreadyActive
	}

	u.active = true
	u.updatedAt = now.UTC()

	u.addEvent(event.NewUserActivated(u.id.String(), now))

	return nil
}

func (u *User) ChangePassword(newHash string, now time.Time) error {
	if !u.isValid() {
		return domainErrors.ErrCorruptedUser
	}
	if newHash == "" {
		return domainErrors.ErrInvalidPasswordHash
	}
	if newHash == u.passwordHash {
		return domainErrors.ErrPasswordUnchanged
	}

	u.passwordHash = newHash
	u.updatedAt = now.UTC()

	u.addEvent(event.NewPasswordChanged(u.id.String(), now))

	return nil
}

func (u *User) isValid() bool {
	return !u.id.IsZero() && !u.email.IsZero() && u.passwordHash != "" && !u.role.IsZero()
}

func (u *User) addEvent(evt events.DomainEvent) {
	u.domainEvents = append(u.domainEvents, evt)
}

func (u *User) DomainEvents() []events.DomainEvent {
	cp := make([]events.DomainEvent, len(u.domainEvents))
	copy(cp, u.domainEvents)
	return cp
}

func (u *User) PullEvents() []events.DomainEvent {
	evts := u.DomainEvents()
	u.ClearEvents()
	return evts
}

func (u *User) ClearEvents() {
	u.domainEvents = nil
}

func (u *User) ID() valueobject.UserID {
	return u.id
}

func (u *User) Email() valueobject.Email {
	return u.email
}

func (u *User) PasswordHash() string {
	return u.passwordHash
}

func (u *User) Role() valueobject.Role {
	return u.role
}

func (u *User) Active() bool {
	return u.active
}

func (u *User) CreatedAt() time.Time {
	return u.createdAt
}

func (u *User) UpdatedAt() time.Time {
	return u.updatedAt
}
