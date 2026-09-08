package domain

import (
	"errors"
	"context"

	"github.com/google/uuid"
)

type CartItemInput struct {
	ProductID uuid.UUID
	Quantity  int
}

var ErrInsufficientStock = errors.New("checkout: insufficient stock for one or more products")

// ============================================================
// CheckoutService -- SEKARANG PURE, tanpa OrderRepository.
// ============================================================
// Perhatikan struct-nya cuma punya stockChecker (untuk BACA harga
// & stok, dibutuhkan untuk KONSTRUKSI Order). TIDAK ADA lagi
// orderRepo, TIDAK ADA lagi pemanggilan Save()/ReduceStock() di
// dalam service ini. CheckoutFull cuma MEMBANGUN & MENGEMBALIKAN
// *Order -- semua efek samping tulis (persist order, kurangi stok,
// panggil payment gateway) jadi tanggung jawab CheckoutUseCase.
//
// Konsekuensi: atomicity "checkout = 1 unit kerja" sekarang harus
// dijaga EKSPLISIT di use case (lihat komentar TODO di
// checkout_usecase.go), bukan implisit di sini. Trade-off yang
// disengaja -- lihat diskusi Pendekatan A vs B.
// ============================================================

type CheckoutService struct {
	stockChecker ProductStockChecker
}

func NewCheckoutService(stockChecker ProductStockChecker) *CheckoutService {
	return &CheckoutService{stockChecker: stockChecker}
}

func (s *CheckoutService) CheckoutFull(ctx context.Context, userID uuid.UUID, cartItems []CartItemInput, discountPolicy DiscountPolicy, adminFeePolicy AdminFeePolicy) (*Order, error) {
	orderItems, err := s.buildOrderItems(ctx, cartItems)
	if err != nil {
		return nil, err
	}
	order, err := NewOrder(userID, orderItems, IDR)
	if err != nil {
		return nil, err
	}
	if err := order.ApplyDiscount(discountPolicy); err != nil {
		return nil, err
	}
	if err := order.ApplyAdminFee(adminFeePolicy); err != nil {
		return nil, err
	}
	return order, nil
}

func (s *CheckoutService) Checkout(ctx context.Context, userID uuid.UUID, cartItems []CartItemInput) (*Order, error) {
	return s.CheckoutFull(ctx, userID, cartItems, NoDiscountPolicy{}, NoAdminFee{})
}

func (s *CheckoutService) CheckoutWithDiscount(ctx context.Context, userID uuid.UUID, cartItems []CartItemInput, policy DiscountPolicy) (*Order, error) {
	return s.CheckoutFull(ctx, userID, cartItems, policy, NoAdminFee{})
}

func (s *CheckoutService) buildOrderItems(ctx context.Context, cartItems []CartItemInput) ([]OrderItem, error) {
	orderItems := make([]OrderItem, 0, len(cartItems))
	for _, ci := range cartItems {
		stock, err := s.stockChecker.GetAvailableStock(ctx, ci.ProductID)
		if err != nil {
			return nil, err
		}
		if stock < ci.Quantity {
			return nil, ErrInsufficientStock
		}
		price, name, err := s.stockChecker.GetProductPrice(ctx, ci.ProductID)
		if err != nil {
			return nil, err
		}
		orderItems = append(orderItems, OrderItem{id: uuid.New(), productID: ci.ProductID, productName: name, price: price, quantity: ci.Quantity})
	}
	return orderItems, nil
}
