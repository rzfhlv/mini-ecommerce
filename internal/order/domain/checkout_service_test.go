package domain

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

type mockStockChecker struct {
	stocks map[uuid.UUID]int
	prices map[uuid.UUID]int64
	names  map[uuid.UUID]string
}

func newMockStockChecker() *mockStockChecker {
	return &mockStockChecker{stocks: make(map[uuid.UUID]int), prices: make(map[uuid.UUID]int64), names: make(map[uuid.UUID]string)}
}
func (m *mockStockChecker) GetAvailableStock(ctx context.Context, productID uuid.UUID) (int, error) { return m.stocks[productID], nil }
func (m *mockStockChecker) GetProductPrice(ctx context.Context, productID uuid.UUID) (Money, string, error) {
	money, _ := NewMoney(m.prices[productID], IDR)
	return money, m.names[productID], nil
}
func (m *mockStockChecker) ReduceStock(ctx context.Context, productID uuid.UUID, qty int) error { return nil }

// ============================================================
// PERHATIKAN: test ini TIDAK PERLU mockOrderRepo lagi sama sekali
// -- konsekuensi langsung dari refactor Pendekatan B. CheckoutService
// sekarang cuma butuh stockChecker.
// ============================================================

func TestCheckoutService_Success(t *testing.T) {
	stockChecker := newMockStockChecker()
	productID := uuid.New()
	stockChecker.stocks[productID] = 10
	stockChecker.prices[productID] = 50_000
	stockChecker.names[productID] = "Kemeja Flannel"

	service := NewCheckoutService(stockChecker)
	order, err := service.Checkout(context.Background(), uuid.New(), []CartItemInput{{ProductID: productID, Quantity: 3}})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if order.TotalAmount().Amount() != 150_000 {
		t.Errorf("expected total 150000, got: %d", order.TotalAmount().Amount())
	}
	// CheckoutService TIDAK LAGI mengurangi stok -- itu sekarang
	// tanggung jawab CheckoutUseCase. Stok mock TETAP 10 di sini.
	if stockChecker.stocks[productID] != 10 {
		t.Errorf("expected stock UNCHANGED at 10 (CheckoutService no longer reduces stock), got: %d", stockChecker.stocks[productID])
	}
}

func TestCheckoutService_InsufficientStock(t *testing.T) {
	stockChecker := newMockStockChecker()
	productID := uuid.New()
	stockChecker.stocks[productID] = 2

	service := NewCheckoutService(stockChecker)
	_, err := service.Checkout(context.Background(), uuid.New(), []CartItemInput{{ProductID: productID, Quantity: 5}})

	if err != ErrInsufficientStock {
		t.Errorf("expected ErrInsufficientStock, got: %v", err)
	}
}

func TestCheckoutService_CheckoutFull_WithDiscountAndAdminFee(t *testing.T) {
	stockChecker := newMockStockChecker()
	productID := uuid.New()
	stockChecker.stocks[productID] = 10
	stockChecker.prices[productID] = 100_000
	stockChecker.names[productID] = "Celana Chino"

	service := NewCheckoutService(stockChecker)
	discountPolicy, _ := NewPercentageDiscount(10, "Diskon 10%")
	feePolicy, _ := NewPercentageAdminFee(290, "Fee 2.9%")

	order, err := service.CheckoutFull(context.Background(), uuid.New(), []CartItemInput{{ProductID: productID, Quantity: 1}}, discountPolicy, feePolicy)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	// 100.000 - 10.000 (diskon) + 2.900 (fee) = 92.900
	if order.TotalAmount().Amount() != 92_900 {
		t.Errorf("expected total 92900, got: %d", order.TotalAmount().Amount())
	}
}
