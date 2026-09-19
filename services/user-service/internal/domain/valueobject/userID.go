package valueobject

import (
	"encoding/json"

	domainErrors "github.com/FranciscoHonorat/ordemflow/services/user-service/internal/domain/domain-errors"
	"github.com/google/uuid"
)

type UserID struct {
	id uuid.UUID
}

func NewUserID(id uuid.UUID) (UserID, error) {
	if id == uuid.Nil {
		return UserID{}, domainErrors.ErrInvalidUserID
	}

	return UserID{id: id}, nil
}

func NewUserIDMust(id uuid.UUID) UserID {
	userID, err := NewUserID(id)
	if err != nil {
		panic(err)
	}
	return userID
}

func ParseUserID(idStr string) (UserID, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return UserID{}, domainErrors.ErrInvalidUserID
	}

	return NewUserID(id)
}

func (u UserID) ID() uuid.UUID {
	return u.id
}

func (u UserID) String() string {
	return u.id.String()
}

func (u UserID) Equal(other UserID) bool {
	return u.id == other.id
}

func (u UserID) IsZero() bool {
	return u.id == uuid.Nil
}

func (u UserID) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.id.String())
}

func (u *UserID) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	parsed, err := ParseUserID(value)
	if err != nil {
		return err
	}

	*u = parsed
	return nil
}
