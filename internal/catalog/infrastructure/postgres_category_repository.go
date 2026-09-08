package infrastructure

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"mini-ecommerce/internal/catalog/domain"
)

var _ domain.CategoryRepository = (*PostgresCategoryRepository)(nil)

type PostgresCategoryRepository struct{ db *sql.DB }

func NewPostgresCategoryRepository(db *sql.DB) *PostgresCategoryRepository { return &PostgresCategoryRepository{db: db} }

func (r *PostgresCategoryRepository) List(ctx context.Context) ([]*domain.Category, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, slug FROM categories ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var categories []*domain.Category
	for rows.Next() {
		var id uuid.UUID
		var name, slug string
		if err := rows.Scan(&id, &name, &slug); err != nil {
			return nil, err
		}
		categories = append(categories, domain.ReconstructCategory(id, name, slug))
	}
	return categories, rows.Err()
}
