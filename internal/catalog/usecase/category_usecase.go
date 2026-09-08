package usecase

import (
	"context"

	"mini-ecommerce/internal/catalog/domain"
)

type CategoryResponse struct{ ID, Name, Slug string }

type ListCategoriesUseCase interface{ Execute(ctx context.Context) ([]*CategoryResponse, error) }
type listCategoriesUseCase struct{ repo domain.CategoryRepository }

func NewListCategoriesUseCase(repo domain.CategoryRepository) ListCategoriesUseCase { return &listCategoriesUseCase{repo: repo} }
func (uc *listCategoriesUseCase) Execute(ctx context.Context) ([]*CategoryResponse, error) {
	categories, err := uc.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*CategoryResponse, 0, len(categories))
	for _, c := range categories {
		result = append(result, &CategoryResponse{ID: c.ID().String(), Name: c.Name(), Slug: c.Slug()})
	}
	return result, nil
}
