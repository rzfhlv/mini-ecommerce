package domain

// NoDiscountPolicy -- Null Object Pattern. "Tidak ada diskon" adalah
// VARIAN dari DiscountPolicy, bukan percabangan if/else terpisah.
type NoDiscountPolicy struct{}

func (NoDiscountPolicy) Apply(subtotal Money) (Money, error) { return ZeroMoney(subtotal.Currency()), nil }
func (NoDiscountPolicy) Describe() string                    { return "" }
