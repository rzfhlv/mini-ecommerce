package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"mini-ecommerce/internal/catalog/domain"
)

var ErrInvalidProductID = errors.New("usecase: invalid product id")

type CreateProductRequest struct {
	CategoryID  string `json:"category_id" validate:"required,uuid"`
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	Price       int64  `json:"price" validate:"gte=0"`
	Stock       int    `json:"stock" validate:"gte=0"`
	Size        string `json:"size"`
	ImageURL    string `json:"image_url"`
}
type UpdateProductRequest struct {
	ID          string `json:"-"`
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	Price       int64  `json:"price" validate:"gte=0"`
	Size        string `json:"size"`
	ImageURL    string `json:"image_url"`
}
type ProductResponse struct {
	ID, CategoryID, Name, Description string
	Price                              int64
	Stock                              int
	Size, ImageURL                     string
}
type ListProductsResponse struct {
	Products    []*ProductResponse `json:"products"`
	Total       int64              `json:"total"`
	Page, Limit int
}

type CreateProductUseCase interface{ Execute(ctx context.Context, req CreateProductRequest) (*ProductResponse, error) }
type createProductUseCase struct{ repo domain.ProductRepository }

func NewCreateProductUseCase(repo domain.ProductRepository) CreateProductUseCase { return &createProductUseCase{repo: repo} }
func (uc *createProductUseCase) Execute(ctx context.Context, req CreateProductRequest) (*ProductResponse, error) {
	categoryID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		return nil, domain.ErrCategoryRequired
	}
	price, err := domain.NewPrice(req.Price)
	if err != nil {
		return nil, err
	}
	product, err := domain.NewProduct(categoryID, req.Name, req.Description, price, req.Stock, req.Size, req.ImageURL)
	if err != nil {
		return nil, err
	}
	if err := uc.repo.Save(ctx, product); err != nil {
		return nil, err
	}
	return toProductResponse(product), nil
}

type UpdateProductUseCase interface{ Execute(ctx context.Context, req UpdateProductRequest) (*ProductResponse, error) }
type updateProductUseCase struct{ repo domain.ProductRepository }

func NewUpdateProductUseCase(repo domain.ProductRepository) UpdateProductUseCase { return &updateProductUseCase{repo: repo} }
func (uc *updateProductUseCase) Execute(ctx context.Context, req UpdateProductRequest) (*ProductResponse, error) {
	id, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, ErrInvalidProductID
	}
	product, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	price, err := domain.NewPrice(req.Price)
	if err != nil {
		return nil, err
	}
	if err := product.UpdateDetails(req.Name, req.Description, price, req.Size, req.ImageURL); err != nil {
		return nil, err
	}
	if err := uc.repo.Update(ctx, product); err != nil {
		return nil, err
	}
	return toProductResponse(product), nil
}

type DeleteProductUseCase interface{ Execute(ctx context.Context, id string) error }
type deleteProductUseCase struct{ repo domain.ProductRepository }

func NewDeleteProductUseCase(repo domain.ProductRepository) DeleteProductUseCase { return &deleteProductUseCase{repo: repo} }
func (uc *deleteProductUseCase) Execute(ctx context.Context, id string) error {
	productID, err := uuid.Parse(id)
	if err != nil {
		return ErrInvalidProductID
	}
	return uc.repo.Delete(ctx, productID)
}

type GetProductUseCase interface{ Execute(ctx context.Context, id string) (*ProductResponse, error) }
type getProductUseCase struct{ repo domain.ProductRepository }

func NewGetProductUseCase(repo domain.ProductRepository) GetProductUseCase { return &getProductUseCase{repo: repo} }
func (uc *getProductUseCase) Execute(ctx context.Context, id string) (*ProductResponse, error) {
	productID, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrInvalidProductID
	}
	product, err := uc.repo.FindByID(ctx, productID)
	if err != nil {
		return nil, err
	}
	return toProductResponse(product), nil
}

type ListProductsUseCase interface {
	Execute(ctx context.Context, categorySlug string, page, limit int) (*ListProductsResponse, error)
}
type listProductsUseCase struct{ repo domain.ProductRepository }

func NewListProductsUseCase(repo domain.ProductRepository) ListProductsUseCase { return &listProductsUseCase{repo: repo} }
func (uc *listProductsUseCase) Execute(ctx context.Context, categorySlug string, page, limit int) (*ListProductsResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit
	products, total, err := uc.repo.List(ctx, categorySlug, offset, limit)
	if err != nil {
		return nil, err
	}
	responses := make([]*ProductResponse, 0, len(products))
	for _, p := range products {
		responses = append(responses, toProductResponse(p))
	}
	return &ListProductsResponse{Products: responses, Total: total, Page: page, Limit: limit}, nil
}

func toProductResponse(p *domain.Product) *ProductResponse {
	return &ProductResponse{ID: p.ID().String(), CategoryID: p.CategoryID().String(), Name: p.Name(), Description: p.Description(), Price: p.Price().Amount(), Stock: p.Stock(), Size: p.Size(), ImageURL: p.ImageURL()}
}
