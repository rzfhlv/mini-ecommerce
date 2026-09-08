package domain

import "errors"

var ErrDiscountExceedsSubtotalPolicy = errors.New("discount: discount amount cannot exceed subtotal")

// DiscountPolicy -- Strategy Pattern. Order/CheckoutService tidak
// pernah tahu jenis promo apa yang dipakai.
type DiscountPolicy interface {
	Apply(subtotal Money) (Money, error)
	Describe() string
}
