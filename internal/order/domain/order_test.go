package domain

import (
	"testing"

	"github.com/google/uuid"
)

func newTestOrderItem(price int64, qty int) OrderItem {
	money, _ := NewMoney(price, IDR)
	return OrderItem{id: uuid.New(), productID: uuid.New(), price: money, quantity: qty}
}

func TestNewOrder_CalculatesTotalCorrectly(t *testing.T) {
	items := []OrderItem{newTestOrderItem(50_000, 2), newTestOrderItem(30_000, 1)}
	order, err := NewOrder(uuid.New(), items, IDR)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if order.TotalAmount().Amount() != 130_000 {
		t.Errorf("expected total 130000, got: %d", order.TotalAmount().Amount())
	}
}

func TestNewOrder_RejectsEmptyItems(t *testing.T) {
	_, err := NewOrder(uuid.New(), []OrderItem{}, IDR)
	if err != ErrOrderEmpty {
		t.Errorf("expected ErrOrderEmpty, got: %v", err)
	}
}

func TestOrder_MarkAsPaid_OnlyFromPending(t *testing.T) {
	order, _ := NewOrder(uuid.New(), []OrderItem{newTestOrderItem(10_000, 1)}, IDR)
	if err := order.MarkAsPaid(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if err := order.MarkAsPaid(); err != ErrInvalidStatusChange {
		t.Errorf("expected ErrInvalidStatusChange, got: %v", err)
	}
}

func TestOrder_ApplyDiscount_And_AdminFee_Combined(t *testing.T) {
	order, _ := NewOrder(uuid.New(), []OrderItem{newTestOrderItem(100_000, 1)}, IDR)

	discountPolicy, _ := NewPercentageDiscount(10, "Diskon 10%")
	if err := order.ApplyDiscount(discountPolicy); err != nil {
		t.Fatalf("expected no error applying discount, got: %v", err)
	}
	// subtotal 100.000 - diskon 10.000 = 90.000 (sementara, sebelum fee)
	if order.TotalAmount().Amount() != 90_000 {
		t.Errorf("expected total 90000 after discount only, got: %d", order.TotalAmount().Amount())
	}

	feePolicy, _ := NewPercentageAdminFee(290, "Fee Kartu Kredit 2.9%") // 290 bps = 2.9%
	if err := order.ApplyAdminFee(feePolicy); err != nil {
		t.Fatalf("expected no error applying admin fee, got: %v", err)
	}
	// fee dihitung dari SUBTOTAL (100.000), bukan dari total setelah
	// diskon -- 2.9% x 100.000 = 2.900. Total akhir = 90.000 + 2.900 = 92.900
	if order.AdminFeeAmount().Amount() != 2_900 {
		t.Errorf("expected admin fee 2900, got: %d", order.AdminFeeAmount().Amount())
	}
	if order.TotalAmount().Amount() != 92_900 {
		t.Errorf("expected final total 92900, got: %d", order.TotalAmount().Amount())
	}
}

func TestOrder_ApplyAdminFee_OrderIndependentOfDiscountOrder(t *testing.T) {
	// Membuktikan recomputeTotal() bikin hasil SAMA walau urutan
	// pemanggilan ApplyAdminFee/ApplyDiscount dibalik.
	orderA, _ := NewOrder(uuid.New(), []OrderItem{newTestOrderItem(100_000, 1)}, IDR)
	discountA, _ := NewPercentageDiscount(10, "Diskon")
	feeA, _ := NewPercentageAdminFee(290, "Fee")
	_ = orderA.ApplyDiscount(discountA)
	_ = orderA.ApplyAdminFee(feeA)

	orderB, _ := NewOrder(uuid.New(), []OrderItem{newTestOrderItem(100_000, 1)}, IDR)
	discountB, _ := NewPercentageDiscount(10, "Diskon")
	feeB, _ := NewPercentageAdminFee(290, "Fee")
	_ = orderB.ApplyAdminFee(feeB)
	_ = orderB.ApplyDiscount(discountB)

	if orderA.TotalAmount().Amount() != orderB.TotalAmount().Amount() {
		t.Errorf("expected same total regardless of order: A=%d B=%d", orderA.TotalAmount().Amount(), orderB.TotalAmount().Amount())
	}
}

func TestOrder_ApplyAdminFee_FixedFee(t *testing.T) {
	order, _ := NewOrder(uuid.New(), []OrderItem{newTestOrderItem(50_000, 1)}, IDR)
	feeMoney, _ := NewMoney(4_000, IDR)
	feePolicy := NewFixedAdminFee(feeMoney, "Biaya Transfer Bank")

	if err := order.ApplyAdminFee(feePolicy); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if order.TotalAmount().Amount() != 54_000 {
		t.Errorf("expected total 54000 (50000+4000), got: %d", order.TotalAmount().Amount())
	}
}

func TestOrder_ApplyAdminFee_NoAdminFeeAddsNothing(t *testing.T) {
	order, _ := NewOrder(uuid.New(), []OrderItem{newTestOrderItem(50_000, 1)}, IDR)
	if err := order.ApplyAdminFee(NoAdminFee{}); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if order.TotalAmount().Amount() != 50_000 {
		t.Errorf("expected total unchanged at 50000, got: %d", order.TotalAmount().Amount())
	}
}

func TestNewPercentageAdminFee_RejectsInvalidBasisPoints(t *testing.T) {
	if _, err := NewPercentageAdminFee(0, "invalid"); err != ErrInvalidBasisPoints {
		t.Errorf("expected ErrInvalidBasisPoints for 0, got: %v", err)
	}
	if _, err := NewPercentageAdminFee(10001, "invalid"); err != ErrInvalidBasisPoints {
		t.Errorf("expected ErrInvalidBasisPoints for >10000, got: %v", err)
	}
}

func TestPaymentMethod_ToAdminFeePolicy(t *testing.T) {
	m, err := NewPaymentMethod("EWALLET")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	policy, err := m.ToAdminFeePolicy()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	subtotal, _ := NewMoney(100_000, IDR)
	fee, _ := policy.Apply(subtotal)
	if fee.Amount() != 0 {
		t.Errorf("expected zero fee for e-wallet, got: %d", fee.Amount())
	}
}

func TestNewPaymentMethod_RejectsUnknown(t *testing.T) {
	_, err := NewPaymentMethod("BITCOIN")
	if err != ErrInvalidPaymentMethod {
		t.Errorf("expected ErrInvalidPaymentMethod, got: %v", err)
	}
}
