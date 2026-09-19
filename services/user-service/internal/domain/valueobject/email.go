package valueobject

import (
	"encoding/json"
	"net/mail"
	"strings"

	domainErrors "github.com/FranciscoHonorat/ordemflow/services/user-service/internal/domain/domain-errors"
)

type Email struct {
	value string
}

func NewEmail(value string) (Email, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return Email{}, domainErrors.ErrInvalidEmail
	}

	if _, err := mail.ParseAddress(normalized); err != nil {
		return Email{}, domainErrors.ErrInvalidEmail
	}

	return Email{value: normalized}, nil
}

func NewEmailMust(value string) Email {
	email, err := NewEmail(value)
	if err != nil {
		panic(err)
	}
	return email
}

func (e Email) String() string {
	return e.value
}

func (e Email) IsZero() bool {
	return e.value == ""
}

func (e Email) Equal(other Email) bool {
	return e.value == other.value
}

func (e Email) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.value)
}

func (e *Email) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	email, err := NewEmail(value)
	if err != nil {
		return err
	}

	*e = email
	return nil
}
