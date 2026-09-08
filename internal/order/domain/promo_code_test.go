package domain

import "testing"

func TestNewPromoCode_ValidPercentage(t *testing.T) {
	promo, err := NewPromoCode("hemat10", "Diskon 10%", DiscountTypePercentage, 10, nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if promo.Code() != "HEMAT10" {
		t.Errorf("expected code normalized to HEMAT10, got: %s", promo.Code())
	}
}

func TestNewPromoCode_RejectsInvalidPercentage(t *testing.T) {
	_, err := NewPromoCode("BAD", "Invalid", DiscountTypePercentage, 150, nil, nil)
	if err != ErrInvalidPercentageVal {
		t.Errorf("expected ErrInvalidPercentageVal, got: %v", err)
	}
}

func TestPromoCode_ToDiscountPolicy_Percentage(t *testing.T) {
	promo, _ := NewPromoCode("HEMAT10", "Diskon 10%", DiscountTypePercentage, 10, nil, nil)
	policy, err := promo.ToDiscountPolicy()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	subtotal, _ := NewMoney(200_000, IDR)
	discount, _ := policy.Apply(subtotal)
	if discount.Amount() != 20_000 {
		t.Errorf("expected discount 20000, got: %d", discount.Amount())
	}
}

func TestPromoCode_IsUsable_InactiveFalse(t *testing.T) {
	promo, _ := NewPromoCode("HEMAT10", "Diskon 10%", DiscountTypePercentage, 10, nil, nil)
	promo.Deactivate()
	if promo.IsUsable() {
		t.Errorf("expected inactive promo to be unusable")
	}
}
