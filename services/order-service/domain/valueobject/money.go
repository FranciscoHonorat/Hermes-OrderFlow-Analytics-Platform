package valueobject

import (
	"encoding/json"
	"fmt"

	domainErrors "github.com/FranciscoHonorat/ordemflow/services/order-service/domain/domain-errors"
)

var validCurrencies = map[string]bool{
	"BRL": true,
	"USD": true,
}

type Money struct {
	amount   int64
	currency string
}

func validateMoney(amount int64, currency string) error {
	if amount <= 0 {
		return domainErrors.ErrNegativeAmount
	}
	if !validCurrencies[currency] {
		return domainErrors.ErrInvalidCurrency
	}
	return nil
}

func NewMoney(amount int64, currency string) (Money, error) {
	if err := validateMoney(amount, currency); err != nil {
		return Money{}, err
	}
	return Money{
		amount:   amount,
		currency: currency,
	}, nil
}

func NewMoneyMust(amount int64, currency string) Money {
	m, err := NewMoney(amount, currency)
	if err != nil {
		panic(err)
	}
	return m
}

func NewMoneyFromAmount(amount int64) (Money, error) {
	if err := validateMoney(amount, "BRL"); err != nil {
		return Money{}, err
	}
	return Money{
		amount:   amount,
		currency: "BRL",
	}, nil
}

func (m Money) Validate() error {
	return validateMoney(m.amount, m.currency)
}

func (m Money) Add(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, domainErrors.ErrCurrencyMismatch
	}
	return Money{
		amount:   m.amount + other.amount,
		currency: m.currency,
	}, nil
}

func (m Money) Amount() int64 {
	return m.amount
}

func (m Money) Currency() string {
	return m.currency
}

func (m Money) Equals(other Money) bool {
	return m.amount == other.amount && m.currency == other.currency
}

func (m Money) String() string {
	return fmt.Sprintf("%d %s", m.amount, m.currency)
}

func (m Money) Multiply(quantity int64) (Money, error) {
	if quantity <= 0 {
		return Money{}, domainErrors.ErrInvalidQuantity
	}
	if err := validateMoney(m.amount, m.currency); err != nil {
		return Money{}, err
	}
	return Money{
		amount:   m.amount * quantity,
		currency: m.currency,
	}, nil
}

func (m Money) MarshalJSON() ([]byte, error) {
	auxMoney := struct {
		Amount   int64  `json:"amount"`
		Currency string `json:"currency"`
	}{
		Amount:   m.amount,
		Currency: m.currency,
	}
	return json.Marshal(auxMoney)
}

func (m *Money) UnmarshalJSON(data []byte) error {
	var alias struct {
		Amount   int64  `json:"amount"`
		Currency string `json:"currency"`
	}
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	if alias.Amount != 0 || alias.Currency != "" {
		if alias.Amount < 0 {
			return domainErrors.ErrNegativeAmount
		}
		if !validCurrencies[alias.Currency] {
			return domainErrors.ErrInvalidCurrency
		}
	}
	m.amount = alias.Amount
	m.currency = alias.Currency
	return nil
}
