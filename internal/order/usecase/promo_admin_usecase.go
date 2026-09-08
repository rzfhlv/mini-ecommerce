package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"mini-ecommerce/internal/order/domain"
)

var ErrPromoCodeAlreadyExists = errors.New("usecase: promo code already exists")

type CreatePromoRequest struct {
	Code, Label   string
	DiscountType  string
	DiscountValue int
	MinPurchase   *int64
	ExpiresAt     *time.Time
}
type UpdatePromoRequest struct {
	ID, Label     string
	DiscountValue int
	MinPurchase   *int64
	ExpiresAt     *time.Time
}
type PromoResponse struct {
	ID, Code, Label, DiscountType string
	DiscountValue                int
	MinPurchase                  *int64
	IsActive                     bool
	ExpiresAt                    *time.Time
}

func toPromoResponse(p *domain.PromoCode) *PromoResponse {
	return &PromoResponse{ID: p.ID().String(), Code: p.Code(), Label: p.Label(), DiscountType: string(p.DiscountType()), DiscountValue: p.DiscountValue(), MinPurchase: p.MinPurchase(), IsActive: p.IsActive(), ExpiresAt: p.ExpiresAt()}
}

type CreatePromoUseCase interface{ Execute(ctx context.Context, req CreatePromoRequest) (*PromoResponse, error) }
type createPromoUseCase struct{ repo domain.PromoCodeAdminRepository }

func NewCreatePromoUseCase(repo domain.PromoCodeAdminRepository) CreatePromoUseCase { return &createPromoUseCase{repo: repo} }
func (uc *createPromoUseCase) Execute(ctx context.Context, req CreatePromoRequest) (*PromoResponse, error) {
	exists, err := uc.repo.ExistsByCode(ctx, req.Code)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrPromoCodeAlreadyExists
	}
	promo, err := domain.NewPromoCode(req.Code, req.Label, domain.DiscountType(req.DiscountType), req.DiscountValue, req.MinPurchase, req.ExpiresAt)
	if err != nil {
		return nil, err
	}
	if err := uc.repo.Save(ctx, promo); err != nil {
		return nil, err
	}
	return toPromoResponse(promo), nil
}

type UpdatePromoUseCase interface{ Execute(ctx context.Context, req UpdatePromoRequest) (*PromoResponse, error) }
type updatePromoUseCase struct{ repo domain.PromoCodeAdminRepository }

func NewUpdatePromoUseCase(repo domain.PromoCodeAdminRepository) UpdatePromoUseCase { return &updatePromoUseCase{repo: repo} }
func (uc *updatePromoUseCase) Execute(ctx context.Context, req UpdatePromoRequest) (*PromoResponse, error) {
	id, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, ErrInvalidPromoID
	}
	promo, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := promo.UpdateDetails(req.Label, req.DiscountValue, req.MinPurchase, req.ExpiresAt); err != nil {
		return nil, err
	}
	if err := uc.repo.Update(ctx, promo); err != nil {
		return nil, err
	}
	return toPromoResponse(promo), nil
}

type SetPromoActiveUseCase interface{ Execute(ctx context.Context, id string, active bool) (*PromoResponse, error) }
type setPromoActiveUseCase struct{ repo domain.PromoCodeAdminRepository }

func NewSetPromoActiveUseCase(repo domain.PromoCodeAdminRepository) SetPromoActiveUseCase { return &setPromoActiveUseCase{repo: repo} }
func (uc *setPromoActiveUseCase) Execute(ctx context.Context, id string, active bool) (*PromoResponse, error) {
	promoID, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrInvalidPromoID
	}
	promo, err := uc.repo.FindByID(ctx, promoID)
	if err != nil {
		return nil, err
	}
	if active {
		promo.Activate()
	} else {
		promo.Deactivate()
	}
	if err := uc.repo.Update(ctx, promo); err != nil {
		return nil, err
	}
	return toPromoResponse(promo), nil
}

type DeletePromoUseCase interface{ Execute(ctx context.Context, id string) error }
type deletePromoUseCase struct{ repo domain.PromoCodeAdminRepository }

func NewDeletePromoUseCase(repo domain.PromoCodeAdminRepository) DeletePromoUseCase { return &deletePromoUseCase{repo: repo} }
func (uc *deletePromoUseCase) Execute(ctx context.Context, id string) error {
	promoID, err := uuid.Parse(id)
	if err != nil {
		return ErrInvalidPromoID
	}
	return uc.repo.Delete(ctx, promoID)
}

type GetPromoUseCase interface{ Execute(ctx context.Context, id string) (*PromoResponse, error) }
type getPromoUseCase struct{ repo domain.PromoCodeAdminRepository }

func NewGetPromoUseCase(repo domain.PromoCodeAdminRepository) GetPromoUseCase { return &getPromoUseCase{repo: repo} }
func (uc *getPromoUseCase) Execute(ctx context.Context, id string) (*PromoResponse, error) {
	promoID, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrInvalidPromoID
	}
	promo, err := uc.repo.FindByID(ctx, promoID)
	if err != nil {
		return nil, err
	}
	return toPromoResponse(promo), nil
}

type ListPromoUseCase interface{ Execute(ctx context.Context, page, limit int) ([]*PromoResponse, error) }
type listPromoUseCase struct{ repo domain.PromoCodeAdminRepository }

func NewListPromoUseCase(repo domain.PromoCodeAdminRepository) ListPromoUseCase { return &listPromoUseCase{repo: repo} }
func (uc *listPromoUseCase) Execute(ctx context.Context, page, limit int) ([]*PromoResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	promos, err := uc.repo.List(ctx, (page-1)*limit, limit)
	if err != nil {
		return nil, err
	}
	responses := make([]*PromoResponse, 0, len(promos))
	for _, p := range promos {
		responses = append(responses, toPromoResponse(p))
	}
	return responses, nil
}
