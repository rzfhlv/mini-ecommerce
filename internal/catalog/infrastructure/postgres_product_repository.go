package infrastructure

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"mini-ecommerce/internal/catalog/domain"
)

var _ domain.ProductRepository = (*PostgresProductRepository)(nil)

type scanRow interface{ Scan(dest ...interface{}) error }

type PostgresProductRepository struct{ db *sql.DB }

func NewPostgresProductRepository(db *sql.DB) *PostgresProductRepository { return &PostgresProductRepository{db: db} }

func (r *PostgresProductRepository) Save(ctx context.Context, p *domain.Product) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO products (id, category_id, name, description, price, stock, size, image_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)
	`, p.ID(), p.CategoryID(), p.Name(), p.Description(), p.Price().Amount(), p.Stock(), p.Size(), p.ImageURL(), p.CreatedAt())
	return err
}

func (r *PostgresProductRepository) Update(ctx context.Context, p *domain.Product) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE products SET name=$1, description=$2, price=$3, stock=$4, size=$5, image_url=$6, updated_at=now() WHERE id=$7
	`, p.Name(), p.Description(), p.Price().Amount(), p.Stock(), p.Size(), p.ImageURL(), p.ID())
	return err
}

func (r *PostgresProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM products WHERE id = $1`, id)
	return err
}

func (r *PostgresProductRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, category_id, name, description, price, stock, size, image_url, created_at, updated_at FROM products WHERE id = $1`, id)
	return scanProduct(row)
}

func (r *PostgresProductRepository) List(ctx context.Context, categorySlug string, offset, limit int) ([]*domain.Product, int64, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT p.id, p.category_id, p.name, p.description, p.price, p.stock, p.size, p.image_url, p.created_at, p.updated_at
		FROM products p JOIN categories c ON c.id = p.category_id
		WHERE $1 = '' OR c.slug = $1 ORDER BY p.created_at DESC LIMIT $2 OFFSET $3
	`, categorySlug, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var products []*domain.Product
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, 0, err
		}
		products = append(products, p)
	}
	var total int64
	err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM products p JOIN categories c ON c.id = p.category_id WHERE $1 = '' OR c.slug = $1`, categorySlug).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	return products, total, rows.Err()
}

func scanProduct(row scanRow) (*domain.Product, error) {
	var (
		id, categoryID       uuid.UUID
		name, description    string
		price                int64
		stock                int
		size, imageURL       sql.NullString
		createdAt, updatedAt sql.NullTime
	)
	if err := row.Scan(&id, &categoryID, &name, &description, &price, &stock, &size, &imageURL, &createdAt, &updatedAt); err != nil {
		return nil, err
	}
	priceVO, err := domain.NewPrice(price)
	if err != nil {
		return nil, err
	}
	return domain.ReconstructProduct(id, categoryID, name, description, priceVO, stock, size.String, imageURL.String, createdAt.Time, updatedAt.Time), nil
}
