package usecase

import (
	"context"

	"github.com/google/uuid"
	"mini-ecommerce/internal/cart/domain"
)

type AddItemRequest struct {
	UserID    string `json:"-"`
	ProductID string `json:"product_id" validate:"required,uuid"`
	Quantity  int    `json:"quantity" validate:"required,gt=0"`
}
type UpdateItemQuantityRequest struct {
	UserID, ProductID string
	Quantity          int `json:"quantity" validate:"required,gt=0"`
}
type CartItemResponse struct {
	ProductID, ProductName string
	Price                  int64
	Quantity               int
	Subtotal               int64
}
type CartResponse struct {
	Items []CartItemResponse `json:"items"`
	Total int64              `json:"total"`
}

func toCartResponse(cart *domain.Cart) *CartResponse {
	items := make([]CartItemResponse, 0, len(cart.Items()))
	for _, i := range cart.Items() {
		items = append(items, CartItemResponse{ProductID: i.ProductID().String(), ProductName: i.ProductName(), Price: i.PriceSnapshot(), Quantity: i.Quantity(), Subtotal: i.Subtotal()})
	}
	return &CartResponse{Items: items, Total: cart.Total()}
}

type AddItemUseCase interface{ Execute(ctx context.Context, req AddItemRequest) (*CartResponse, error) }
type addItemUseCase struct {
	repo   domain.CartRepository
	pricer domain.ProductPricer
}

func NewAddItemUseCase(repo domain.CartRepository, pricer domain.ProductPricer) AddItemUseCase { return &addItemUseCase{repo: repo, pricer: pricer} }
func (uc *addItemUseCase) Execute(ctx context.Context, req AddItemRequest) (*CartResponse, error) {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, ErrInvalidUserID
	}
	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		return nil, ErrInvalidProductID
	}
	price, name, err := uc.pricer.GetPriceAndName(ctx, productID)
	if err != nil {
		return nil, err
	}
	cart, err := uc.repo.FindOrCreateByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if err := cart.AddItem(productID, name, price, req.Quantity); err != nil {
		return nil, err
	}
	if err := uc.repo.Save(ctx, cart); err != nil {
		return nil, err
	}
	return toCartResponse(cart), nil
}

type UpdateItemQuantityUseCase interface{ Execute(ctx context.Context, req UpdateItemQuantityRequest) (*CartResponse, error) }
type updateItemQuantityUseCase struct{ repo domain.CartRepository }

func NewUpdateItemQuantityUseCase(repo domain.CartRepository) UpdateItemQuantityUseCase { return &updateItemQuantityUseCase{repo: repo} }
func (uc *updateItemQuantityUseCase) Execute(ctx context.Context, req UpdateItemQuantityRequest) (*CartResponse, error) {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, ErrInvalidUserID
	}
	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		return nil, ErrInvalidProductID
	}
	cart, err := uc.repo.FindOrCreateByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if err := cart.UpdateItemQuantity(productID, req.Quantity); err != nil {
		return nil, err
	}
	if err := uc.repo.Save(ctx, cart); err != nil {
		return nil, err
	}
	return toCartResponse(cart), nil
}

type RemoveItemUseCase interface{ Execute(ctx context.Context, userID, productID string) (*CartResponse, error) }
type removeItemUseCase struct{ repo domain.CartRepository }

func NewRemoveItemUseCase(repo domain.CartRepository) RemoveItemUseCase { return &removeItemUseCase{repo: repo} }
func (uc *removeItemUseCase) Execute(ctx context.Context, userIDStr, productIDStr string) (*CartResponse, error) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, ErrInvalidUserID
	}
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		return nil, ErrInvalidProductID
	}
	cart, err := uc.repo.FindOrCreateByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if err := cart.RemoveItem(productID); err != nil {
		return nil, err
	}
	if err := uc.repo.Save(ctx, cart); err != nil {
		return nil, err
	}
	return toCartResponse(cart), nil
}

type GetCartUseCase interface{ Execute(ctx context.Context, userID string) (*CartResponse, error) }
type getCartUseCase struct{ repo domain.CartRepository }

func NewGetCartUseCase(repo domain.CartRepository) GetCartUseCase { return &getCartUseCase{repo: repo} }
func (uc *getCartUseCase) Execute(ctx context.Context, userIDStr string) (*CartResponse, error) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, ErrInvalidUserID
	}
	cart, err := uc.repo.FindOrCreateByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return toCartResponse(cart), nil
}

type ClearCartUseCase interface{ Execute(ctx context.Context, userID string) error }
type clearCartUseCase struct{ repo domain.CartRepository }

func NewClearCartUseCase(repo domain.CartRepository) ClearCartUseCase { return &clearCartUseCase{repo: repo} }
func (uc *clearCartUseCase) Execute(ctx context.Context, userIDStr string) error {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return ErrInvalidUserID
	}
	cart, err := uc.repo.FindOrCreateByUserID(ctx, userID)
	if err != nil {
		return err
	}
	cart.Clear()
	return uc.repo.Save(ctx, cart)
}
