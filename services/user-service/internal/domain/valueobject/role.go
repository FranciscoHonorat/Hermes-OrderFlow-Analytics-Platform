package valueobject

import (
	"encoding/json"

	domainErrors "github.com/FranciscoHonorat/ordemflow/services/user-service/internal/domain/domain-errors"
)

const (
	RoleNameAdmin = "admin"
	RoleNameUser  = "user"
)

type Role struct {
	value string
}

func NewRole(value string) (Role, error) {
	switch value {
	case RoleNameAdmin, RoleNameUser:
		return Role{value: value}, nil
	default:
		return Role{}, domainErrors.ErrInvalidRole
	}
}

func NewRoleMust(value string) Role {
	role, err := NewRole(value)
	if err != nil {
		panic(err)
	}
	return role
}

func (r Role) String() string {
	return r.value
}

func (r Role) IsZero() bool {
	return r.value == ""
}

func (r Role) IsAdmin() bool {
	return r.value == RoleNameAdmin
}

func (r Role) Equal(other Role) bool {
	return r.value == other.value
}

func (r Role) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.value)
}

func (r *Role) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	role, err := NewRole(value)
	if err != nil {
		return err
	}

	*r = role
	return nil
}
