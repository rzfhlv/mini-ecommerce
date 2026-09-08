package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type DiscountType string

const (
	DiscountTypePercentage  DiscountType = "PERCENTAGE"
	DiscountTypeFixedAmount DiscountType = "FIXED_AMOUNT"
)

func (t DiscountType) IsValid() bool { return t == DiscountTypePercentage || t == DiscountTypeFixedAmount }

var (
	ErrEmptyPromoCode       = errors.New("promo: code cannot be empty")
	ErrInvalidDiscountType  = errors.New("promo: invalid discount type")
	ErrInvalidPercentageVal = errors.New("promo: percentage must be between 1 and 100")
	ErrInvalidFixedAmount   = errors.New("promo: fixed amount must be greater than 0")
	ErrPromoAlreadyExpired  = errors.New("promo: expiry date must be in the future")
)

type PromoCode struct {
	id                    uuid.UUID
	code, label           string
	discountType          DiscountType
	discountValue         int
	minPurchase           *int64
	isActive              bool
	expiresAt             *time.Time
	createdAt, updatedAt  time.Time
}

func NewPromoCode(code, label string, discountType DiscountType, discountValue int, minPurchase *int64, expiresAt *time.Time) (*PromoCode, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return nil, ErrEmptyPromoCode
	}
	if !discountType.IsValid() {
		return nil, ErrInvalidDiscountType
	}
	if discountType == DiscountTypePercentage && (discountValue <= 0 || discountValue > 100) {
		return nil, ErrInvalidPercentageVal
	}
	if discountType == DiscountTypeFixedAmount && discountValue <= 0 {
		return nil, ErrInvalidFixedAmount
	}
	if expiresAt != nil && expiresAt.Before(time.Now()) {
		return nil, ErrPromoAlreadyExpired
	}
	now := time.Now()
	return &PromoCode{id: uuid.New(), code: code, label: label, discountType: discountType, discountValue: discountValue, minPurchase: minPurchase, isActive: true, expiresAt: expiresAt, createdAt: now, updatedAt: now}, nil
}

func ReconstructPromoCode(id uuid.UUID, code, label string, discountType DiscountType, discountValue int, minPurchase *int64, isActive bool, expiresAt *time.Time, createdAt, updatedAt time.Time) *PromoCode {
	return &PromoCode{id: id, code: code, label: label, discountType: discountType, discountValue: discountValue, minPurchase: minPurchase, isActive: isActive, expiresAt: expiresAt, createdAt: createdAt, updatedAt: updatedAt}
}

func (p *PromoCode) UpdateDetails(label string, discountValue int, minPurchase *int64, expiresAt *time.Time) error {
	if p.discountType == DiscountTypePercentage && (discountValue <= 0 || discountValue > 100) {
		return ErrInvalidPercentageVal
	}
	if p.discountType == DiscountTypeFixedAmount && discountValue <= 0 {
		return ErrInvalidFixedAmount
	}
	if expiresAt != nil && expiresAt.Before(time.Now()) {
		return ErrPromoAlreadyExpired
	}
	p.label, p.discountValue, p.minPurchase, p.expiresAt = label, discountValue, minPurchase, expiresAt
	p.updatedAt = time.Now()
	return nil
}

func (p *PromoCode) Activate()   { p.isActive = true; p.updatedAt = time.Now() }
func (p *PromoCode) Deactivate() { p.isActive = false; p.updatedAt = time.Now() }

func (p *PromoCode) ID() uuid.UUID              { return p.id }
func (p *PromoCode) Code() string               { return p.code }
func (p *PromoCode) Label() string              { return p.label }
func (p *PromoCode) DiscountType() DiscountType { return p.discountType }
func (p *PromoCode) DiscountValue() int         { return p.discountValue }
func (p *PromoCode) MinPurchase() *int64        { return p.minPurchase }
func (p *PromoCode) IsActive() bool             { return p.isActive }
func (p *PromoCode) ExpiresAt() *time.Time      { return p.expiresAt }
func (p *PromoCode) CreatedAt() time.Time       { return p.createdAt }

func (p *PromoCode) IsUsable() bool {
	if !p.isActive {
		return false
	}
	if p.expiresAt != nil && p.expiresAt.Before(time.Now()) {
		return false
	}
	return true
}

// ToDiscountPolicy -- titik pertemuan dunia CRUD (PromoCode) dan
// dunia checkout (DiscountPolicy). SATU-SATUNYA tempat aturan
// "tipe promo -> strategy" hidup -- infrastructure layer TIDAK
// PERLU switch statement lagi.
func (p *PromoCode) ToDiscountPolicy() (DiscountPolicy, error) {
	var policy DiscountPolicy
	var err error
	switch p.discountType {
	case DiscountTypePercentage:
		policy, err = NewPercentageDiscount(p.discountValue, p.label)
	case DiscountTypeFixedAmount:
		var amount Money
		amount, err = NewMoney(int64(p.discountValue), IDR)
		if err == nil {
			policy = NewFixedAmountDiscount(amount, p.label)
		}
	default:
		return nil, ErrInvalidDiscountType
	}
	if err != nil {
		return nil, err
	}
	if p.minPurchase != nil {
		minMoney, err := NewMoney(*p.minPurchase, IDR)
		if err != nil {
			return nil, err
		}
		policy = NewMinPurchaseDiscount(policy, minMoney)
	}
	return policy, nil
}
