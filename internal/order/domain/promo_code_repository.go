package domain

import (
	"context"
	"errors"
)

var ErrPromoCodeNotFound = errors.New("promo: code not found or expired")

// PromoCodeRepository -- dipakai sisi CHECKOUT (baca via kode promo).
type PromoCodeRepository interface {
	FindPolicy(ctx context.Context, code string) (DiscountPolicy, error)
}
