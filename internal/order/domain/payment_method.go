package domain

import "errors"

var ErrInvalidPaymentMethod = errors.New("payment method: unknown value")

type PaymentMethod string

const (
	PaymentMethodBankTransfer PaymentMethod = "BANK_TRANSFER"
	PaymentMethodCreditCard   PaymentMethod = "CREDIT_CARD"
	PaymentMethodEWallet      PaymentMethod = "EWALLET"
)

func NewPaymentMethod(raw string) (PaymentMethod, error) {
	switch PaymentMethod(raw) {
	case PaymentMethodBankTransfer, PaymentMethodCreditCard, PaymentMethodEWallet:
		return PaymentMethod(raw), nil
	default:
		return "", ErrInvalidPaymentMethod
	}
}

// ToAdminFeePolicy -- SATU-SATUNYA tempat aturan "metode X kena fee
// Y" hidup. HARDCODE (bukan CRUD/database) karena perubahannya jarang
// -- beda dengan PromoCode yang memang butuh diubah kapan saja oleh
// tim marketing tanpa deploy ulang.
func (m PaymentMethod) ToAdminFeePolicy() (AdminFeePolicy, error) {
	switch m {
	case PaymentMethodBankTransfer:
		fee, err := NewMoney(4_000, IDR)
		if err != nil {
			return nil, err
		}
		return NewFixedAdminFee(fee, "Biaya Admin Transfer Bank"), nil
	case PaymentMethodCreditCard:
		return NewPercentageAdminFee(290, "Biaya Admin Kartu Kredit (2.9%)")
	case PaymentMethodEWallet:
		return NoAdminFee{}, nil
	default:
		return nil, ErrInvalidPaymentMethod
	}
}
