package usecase

import (
	"context"

	"github.com/google/uuid"
	"mini-ecommerce/internal/order/domain"
)

type EventPublisher interface {
	Publish(ctx context.Context, event domain.DomainEvent)
}

type CheckoutItemRequest struct {
	ProductID string `json:"product_id" validate:"required,uuid"`
	Quantity  int    `json:"quantity" validate:"required,gt=0"`
}
type CheckoutRequest struct {
	UserID        string                `json:"-"`
	Items         []CheckoutItemRequest `json:"items" validate:"required,min=1,dive"`
	PromoCode     string                `json:"promo_code,omitempty"`
	PaymentMethod string                `json:"payment_method" validate:"required"`
}
type OrderItemResponse struct {
	ProductID, ProductName string
	Price                  int64
	Quantity               int
	Subtotal               int64
}
type OrderResponse struct {
	ID, Status             string
	Subtotal               int64
	DiscountAmount         int64
	DiscountLabel          string
	AdminFeeAmount         int64
	AdminFeeLabel          string
	TotalAmount            int64
	Items                  []OrderItemResponse
	PaymentRedirectURL     string
}

type CheckoutUseCase interface {
	Execute(ctx context.Context, req CheckoutRequest) (*OrderResponse, error)
}

type checkoutUseCase struct {
	checkoutService *domain.CheckoutService
	stockChecker    domain.ProductStockChecker // dibutuhkan DI SINI sekarang untuk ReduceStock (dulu di dalam CheckoutService)
	orderRepo       domain.OrderRepository
	promoRepo       domain.PromoCodeRepository
	paymentGateway  domain.PaymentGateway
	eventPublisher  EventPublisher
}

func NewCheckoutUseCase(checkoutService *domain.CheckoutService, stockChecker domain.ProductStockChecker, orderRepo domain.OrderRepository, promoRepo domain.PromoCodeRepository, paymentGateway domain.PaymentGateway, publisher EventPublisher) CheckoutUseCase {
	return &checkoutUseCase{checkoutService: checkoutService, stockChecker: stockChecker, orderRepo: orderRepo, promoRepo: promoRepo, paymentGateway: paymentGateway, eventPublisher: publisher}
}

func (uc *checkoutUseCase) Execute(ctx context.Context, req CheckoutRequest) (*OrderResponse, error) {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, ErrInvalidUserID
	}

	items := make([]domain.CartItemInput, 0, len(req.Items))
	for _, i := range req.Items {
		productID, err := uuid.Parse(i.ProductID)
		if err != nil {
			return nil, ErrInvalidProductID
		}
		items = append(items, domain.CartItemInput{ProductID: productID, Quantity: i.Quantity})
	}

	paymentMethod, err := domain.NewPaymentMethod(req.PaymentMethod)
	if err != nil {
		return nil, err
	}
	adminFeePolicy, err := paymentMethod.ToAdminFeePolicy()
	if err != nil {
		return nil, err
	}

	var discountPolicy domain.DiscountPolicy = domain.NoDiscountPolicy{}
	if req.PromoCode != "" {
		discountPolicy, err = uc.promoRepo.FindPolicy(ctx, req.PromoCode)
		if err != nil {
			return nil, err
		}
	}

	// 1. Domain Service cuma MEMBANGUN order -- belum ada efek samping.
	order, err := uc.checkoutService.CheckoutFull(ctx, userID, items, discountPolicy, adminFeePolicy)
	if err != nil {
		return nil, err
	}

	// 2. SEMUA efek samping tulis diorkestrasi DI SINI, berurutan,
	//    eksplisit -- ini yang berubah sejak refactor Pendekatan B.
	for _, item := range order.Items() {
		if err := uc.stockChecker.ReduceStock(ctx, item.ProductID(), item.Quantity()); err != nil {
			// TODO produksi: kalau gagal di tengah loop, item sebelumnya
			// sudah terlanjur dikurangi -- butuh compensating action atau
			// ReduceStock versi batch (1 panggilan atomik untuk semua item).
			return nil, err
		}
	}

	if err := uc.orderRepo.Save(ctx, order); err != nil {
		return nil, err
	}

	initiation, err := uc.paymentGateway.Initiate(ctx, order.ID(), order.TotalAmount())
	if err != nil {
		return nil, err
	}
	if err := order.SetPaymentReference(initiation.PaymentReference); err != nil {
		return nil, err
	}
	if err := uc.orderRepo.Update(ctx, order); err != nil {
		return nil, err
	}

	for _, event := range order.PullEvents() {
		uc.eventPublisher.Publish(ctx, event)
	}

	resp := toOrderResponse(order)
	resp.PaymentRedirectURL = initiation.RedirectURL
	return resp, nil
}

func toOrderResponse(order *domain.Order) *OrderResponse {
	items := make([]OrderItemResponse, 0, len(order.Items()))
	for _, item := range order.Items() {
		items = append(items, OrderItemResponse{ProductID: item.ProductID().String(), ProductName: item.ProductName(), Price: item.Price().Amount(), Quantity: item.Quantity(), Subtotal: item.Subtotal().Amount()})
	}
	return &OrderResponse{
		ID: order.ID().String(), Status: string(order.Status()),
		Subtotal: order.Subtotal().Amount(), DiscountAmount: order.DiscountAmount().Amount(), DiscountLabel: order.DiscountLabel(),
		AdminFeeAmount: order.AdminFeeAmount().Amount(), AdminFeeLabel: order.AdminFeeLabel(),
		TotalAmount: order.TotalAmount().Amount(), Items: items,
	}
}
