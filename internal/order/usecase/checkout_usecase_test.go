package usecase

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"mini-ecommerce/internal/order/domain"
)

type stubStockChecker struct {
	stock map[uuid.UUID]int
	price map[uuid.UUID]int64
	name  map[uuid.UUID]string
	// reduceCalls MEREKAM setiap panggilan ReduceStock -- ini yang
	// membuktikan efek samping "kurangi stok" sekarang BENAR-BENAR
	// terjadi di use case, bukan lagi di CheckoutService.
	reduceCalls []struct {
		ProductID uuid.UUID
		Qty       int
	}
}

func newStubStockChecker() *stubStockChecker {
	return &stubStockChecker{stock: map[uuid.UUID]int{}, price: map[uuid.UUID]int64{}, name: map[uuid.UUID]string{}}
}
func (s *stubStockChecker) GetAvailableStock(ctx context.Context, productID uuid.UUID) (int, error) { return s.stock[productID], nil }
func (s *stubStockChecker) GetProductPrice(ctx context.Context, productID uuid.UUID) (domain.Money, string, error) {
	m, _ := domain.NewMoney(s.price[productID], domain.IDR)
	return m, s.name[productID], nil
}
func (s *stubStockChecker) ReduceStock(ctx context.Context, productID uuid.UUID, qty int) error {
	s.reduceCalls = append(s.reduceCalls, struct {
		ProductID uuid.UUID
		Qty       int
	}{productID, qty})
	s.stock[productID] -= qty
	return nil
}

type stubOrderRepo struct {
	saved       []*domain.Order
	updateCalls int
}

func (s *stubOrderRepo) Save(ctx context.Context, order *domain.Order) error { s.saved = append(s.saved, order); return nil }
func (s *stubOrderRepo) Update(ctx context.Context, order *domain.Order) error { s.updateCalls++; return nil }
func (s *stubOrderRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	for _, o := range s.saved {
		if o.ID() == id {
			return o, nil
		}
	}
	return nil, nil
}
func (s *stubOrderRepo) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Order, error) { return nil, nil }

type stubPromoRepo struct{ policies map[string]domain.DiscountPolicy }

func (s *stubPromoRepo) FindPolicy(ctx context.Context, code string) (domain.DiscountPolicy, error) {
	p, ok := s.policies[code]
	if !ok {
		return nil, domain.ErrPromoCodeNotFound
	}
	return p, nil
}

type stubPaymentGateway struct{}

func (s *stubPaymentGateway) Initiate(ctx context.Context, orderID uuid.UUID, amount domain.Money) (domain.PaymentInitiation, error) {
	return domain.PaymentInitiation{PaymentReference: "stub-ref", RedirectURL: "https://pay.example/stub"}, nil
}

type stubEventPublisher struct{ published []domain.DomainEvent }

func (s *stubEventPublisher) Publish(ctx context.Context, event domain.DomainEvent) { s.published = append(s.published, event) }

func TestCheckoutUseCase_ReducesStockAndSavesOrder(t *testing.T) {
	stockChecker := newStubStockChecker()
	productID := uuid.New()
	stockChecker.stock[productID] = 10
	stockChecker.price[productID] = 100_000
	stockChecker.name[productID] = "Kemeja"

	orderRepo := &stubOrderRepo{}
	checkoutService := domain.NewCheckoutService(stockChecker)
	uc := NewCheckoutUseCase(checkoutService, stockChecker, orderRepo, &stubPromoRepo{policies: map[string]domain.DiscountPolicy{}}, &stubPaymentGateway{}, &stubEventPublisher{})

	resp, err := uc.Execute(context.Background(), CheckoutRequest{
		UserID: uuid.New().String(), Items: []CheckoutItemRequest{{ProductID: productID.String(), Quantity: 2}}, PaymentMethod: "EWALLET",
	})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	// Buktikan ReduceStock BENAR-BENAR dipanggil DI USE CASE (bukan
	// lagi di CheckoutService, yang sekarang tidak menyentuhnya).
	if len(stockChecker.reduceCalls) != 1 || stockChecker.reduceCalls[0].Qty != 2 {
		t.Fatalf("expected ReduceStock called once with qty=2, got: %+v", stockChecker.reduceCalls)
	}
	if stockChecker.stock[productID] != 8 {
		t.Errorf("expected stock reduced to 8, got: %d", stockChecker.stock[productID])
	}
	// Buktikan order.Save() dipanggil.
	if len(orderRepo.saved) != 1 {
		t.Errorf("expected 1 order saved, got: %d", len(orderRepo.saved))
	}
	// Buktikan orderRepo.Update() dipanggil (untuk simpan payment_reference).
	if orderRepo.updateCalls != 1 {
		t.Errorf("expected 1 Update call for payment reference, got: %d", orderRepo.updateCalls)
	}
	if resp.PaymentRedirectURL != "https://pay.example/stub" {
		t.Errorf("expected redirect URL from stub gateway, got: %s", resp.PaymentRedirectURL)
	}
}

func TestCheckoutUseCase_WithPromoCodeAndAdminFee(t *testing.T) {
	stockChecker := newStubStockChecker()
	productID := uuid.New()
	stockChecker.stock[productID] = 10
	stockChecker.price[productID] = 100_000

	percentagePolicy, _ := domain.NewPercentageDiscount(10, "Diskon 10%")
	orderRepo := &stubOrderRepo{}
	checkoutService := domain.NewCheckoutService(stockChecker)
	uc := NewCheckoutUseCase(checkoutService, stockChecker, orderRepo, &stubPromoRepo{policies: map[string]domain.DiscountPolicy{"HEMAT10": percentagePolicy}}, &stubPaymentGateway{}, &stubEventPublisher{})

	resp, err := uc.Execute(context.Background(), CheckoutRequest{
		UserID: uuid.New().String(), Items: []CheckoutItemRequest{{ProductID: productID.String(), Quantity: 1}},
		PromoCode: "HEMAT10", PaymentMethod: "CREDIT_CARD",
	})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	// subtotal 100.000 - diskon 10% (10.000) + fee kartu kredit 2.9% dari subtotal (2.900) = 92.900
	if resp.DiscountAmount != 10_000 {
		t.Errorf("expected discount 10000, got: %d", resp.DiscountAmount)
	}
	if resp.AdminFeeAmount != 2_900 {
		t.Errorf("expected admin fee 2900, got: %d", resp.AdminFeeAmount)
	}
	if resp.TotalAmount != 92_900 {
		t.Errorf("expected total 92900, got: %d", resp.TotalAmount)
	}
}

func TestCheckoutUseCase_InvalidPaymentMethod(t *testing.T) {
	stockChecker := newStubStockChecker()
	orderRepo := &stubOrderRepo{}
	checkoutService := domain.NewCheckoutService(stockChecker)
	uc := NewCheckoutUseCase(checkoutService, stockChecker, orderRepo, &stubPromoRepo{policies: map[string]domain.DiscountPolicy{}}, &stubPaymentGateway{}, &stubEventPublisher{})

	_, err := uc.Execute(context.Background(), CheckoutRequest{
		UserID: uuid.New().String(), Items: []CheckoutItemRequest{{ProductID: uuid.New().String(), Quantity: 1}}, PaymentMethod: "BITCOIN",
	})

	if err != domain.ErrInvalidPaymentMethod {
		t.Errorf("expected ErrInvalidPaymentMethod, got: %v", err)
	}
	// Karena validasi payment method gagal SEBELUM checkoutService
	// dipanggil, TIDAK ADA stok yang boleh berkurang ataupun order tersimpan.
	if len(orderRepo.saved) != 0 {
		t.Errorf("expected no order saved on invalid payment method, got: %d", len(orderRepo.saved))
	}
}
