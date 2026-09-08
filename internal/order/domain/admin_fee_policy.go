package domain

// AdminFeePolicy -- interface TERPISAH dari DiscountPolicy walau
// shape method-nya identik. Alasan: Ubiquitous Language -- "discount"
// MENGURANGI, "admin fee" MENAMBAH. Interface beda bikin niat
// eksplisit dari signature-nya sendiri, mengurangi risiko salah pasang.
type AdminFeePolicy interface {
	Apply(subtotal Money) (Money, error)
	Describe() string
}

// NoAdminFee -- Null Object, pola sama persis NoDiscountPolicy.
type NoAdminFee struct{}

func (NoAdminFee) Apply(subtotal Money) (Money, error) { return ZeroMoney(subtotal.Currency()), nil }
func (NoAdminFee) Describe() string                    { return "" }
