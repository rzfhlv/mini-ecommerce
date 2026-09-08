package domain

import (
	"errors"
	"fmt"
)

type Currency string

const IDR Currency = "IDR"

type Money struct {
	amount   int64
	currency Currency
}

func NewMoney(amount int64, currency Currency) (Money, error) {
	if amount < 0 {
		return Money{}, errors.New("money: amount cannot be negative")
	}
	if currency == "" {
		return Money{}, errors.New("money: currency is required")
	}
	return Money{amount: amount, currency: currency}, nil
}
func ZeroMoney(currency Currency) Money { return Money{amount: 0, currency: currency} }

func (m Money) Amount() int64      { return m.amount }
func (m Money) Currency() Currency { return m.currency }

func (m Money) Add(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, fmt.Errorf("money: cannot add different currencies (%s vs %s)", m.currency, other.currency)
	}
	return Money{amount: m.amount + other.amount, currency: m.currency}, nil
}
func (m Money) Multiply(qty int) Money { return Money{amount: m.amount * int64(qty), currency: m.currency} }
func (m Money) Equals(other Money) bool { return m.amount == other.amount && m.currency == other.currency }
func (m Money) String() string          { return fmt.Sprintf("%s %d", m.currency, m.amount) }
