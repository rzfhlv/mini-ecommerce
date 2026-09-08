package domain

import "errors"

var ErrInvalidPrice = errors.New("price: amount cannot be negative")

// Price -- Value Object SENDIRI, sengaja TIDAK dibagi dengan Money
// (Order). Lihat penjelasan lengkap di komentar order/domain/money.go
// -- keputusan menghindari Shared Kernel antar bounded context.
type Price struct{ amount int64 }

func NewPrice(amount int64) (Price, error) {
	if amount < 0 {
		return Price{}, ErrInvalidPrice
	}
	return Price{amount: amount}, nil
}
func (p Price) Amount() int64 { return p.amount }
