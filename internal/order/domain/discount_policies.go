package domain

import "errors"

var ErrInvalidPercentage = errors.New("percentage discount: must be between 1 and 100")

type PercentageDiscount struct {
	percentage int
	label      string
}

func NewPercentageDiscount(percentage int, label string) (*PercentageDiscount, error) {
	if percentage <= 0 || percentage > 100 {
		return nil, ErrInvalidPercentage
	}
	return &PercentageDiscount{percentage: percentage, label: label}, nil
}
func (p *PercentageDiscount) Apply(subtotal Money) (Money, error) {
	amount := subtotal.Amount() * int64(p.percentage) / 100
	return NewMoney(amount, subtotal.Currency())
}
func (p *PercentageDiscount) Describe() string { return p.label }

type FixedAmountDiscount struct {
	amount Money
	label  string
}

func NewFixedAmountDiscount(amount Money, label string) *FixedAmountDiscount { return &FixedAmountDiscount{amount: amount, label: label} }
func (f *FixedAmountDiscount) Apply(subtotal Money) (Money, error) {
	if f.amount.Amount() > subtotal.Amount() {
		return NewMoney(subtotal.Amount(), subtotal.Currency())
	}
	return f.amount, nil
}
func (f *FixedAmountDiscount) Describe() string { return f.label }

type MinPurchaseDiscount struct {
	inner   DiscountPolicy
	minimum Money
}

func NewMinPurchaseDiscount(inner DiscountPolicy, minimum Money) *MinPurchaseDiscount { return &MinPurchaseDiscount{inner: inner, minimum: minimum} }
func (m *MinPurchaseDiscount) Apply(subtotal Money) (Money, error) {
	if subtotal.Amount() < m.minimum.Amount() {
		return ZeroMoney(subtotal.Currency()), nil
	}
	return m.inner.Apply(subtotal)
}
func (m *MinPurchaseDiscount) Describe() string { return m.inner.Describe() }
