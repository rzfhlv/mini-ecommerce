package domain

import "errors"

var ErrInvalidBasisPoints = errors.New("admin fee: basis points must be between 1 and 10000")

// PercentageAdminFee pakai BASIS POINTS (per 10.000), bukan persen
// bulat seperti PercentageDiscount -- supaya bisa merepresentasikan
// pecahan seperti 2.9% (kartu kredit) tanpa perlu tipe float.
type PercentageAdminFee struct {
	basisPoints int // 1-10000; 290 = 2.9%
	label       string
}

func NewPercentageAdminFee(basisPoints int, label string) (*PercentageAdminFee, error) {
	if basisPoints <= 0 || basisPoints > 10000 {
		return nil, ErrInvalidBasisPoints
	}
	return &PercentageAdminFee{basisPoints: basisPoints, label: label}, nil
}
func (f *PercentageAdminFee) Apply(subtotal Money) (Money, error) {
	amount := subtotal.Amount() * int64(f.basisPoints) / 10000
	return NewMoney(amount, subtotal.Currency())
}
func (f *PercentageAdminFee) Describe() string { return f.label }

type FixedAdminFee struct {
	amount Money
	label  string
}

func NewFixedAdminFee(amount Money, label string) *FixedAdminFee { return &FixedAdminFee{amount: amount, label: label} }
func (f *FixedAdminFee) Apply(subtotal Money) (Money, error)     { return f.amount, nil }
func (f *FixedAdminFee) Describe() string                        { return f.label }
