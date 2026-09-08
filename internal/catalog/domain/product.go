package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrEmptyProductName  = errors.New("product: name cannot be empty")
	ErrCategoryRequired  = errors.New("product: category is required")
	ErrNegativeStock     = errors.New("product: initial stock cannot be negative")
	ErrInsufficientStock = errors.New("product: insufficient stock")
)

type Product struct {
	id, categoryID       uuid.UUID
	name, description    string
	price                Price
	stock                int
	size, imageURL        string
	createdAt, updatedAt time.Time
}

func NewProduct(categoryID uuid.UUID, name, description string, price Price, stock int, size, imageURL string) (*Product, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrEmptyProductName
	}
	if categoryID == uuid.Nil {
		return nil, ErrCategoryRequired
	}
	if stock < 0 {
		return nil, ErrNegativeStock
	}
	now := time.Now()
	return &Product{id: uuid.New(), categoryID: categoryID, name: name, description: description, price: price, stock: stock, size: size, imageURL: imageURL, createdAt: now, updatedAt: now}, nil
}

func ReconstructProduct(id, categoryID uuid.UUID, name, description string, price Price, stock int, size, imageURL string, createdAt, updatedAt time.Time) *Product {
	return &Product{id: id, categoryID: categoryID, name: name, description: description, price: price, stock: stock, size: size, imageURL: imageURL, createdAt: createdAt, updatedAt: updatedAt}
}

func (p *Product) UpdateDetails(name, description string, price Price, size, imageURL string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrEmptyProductName
	}
	p.name, p.description, p.price, p.size, p.imageURL = name, description, price, size, imageURL
	p.updatedAt = time.Now()
	return nil
}

// DecreaseStock -- SATU-SATUNYA jalan sah mengurangi stok. Aturan
// "tidak boleh minus" hidup DI SINI, bukan di SQL WHERE clause milik
// context lain.
func (p *Product) DecreaseStock(qty int) error {
	if qty > p.stock {
		return ErrInsufficientStock
	}
	p.stock -= qty
	p.updatedAt = time.Now()
	return nil
}
func (p *Product) IncreaseStock(qty int) { p.stock += qty; p.updatedAt = time.Now() }

func (p *Product) ID() uuid.UUID         { return p.id }
func (p *Product) CategoryID() uuid.UUID { return p.categoryID }
func (p *Product) Name() string          { return p.name }
func (p *Product) Description() string   { return p.description }
func (p *Product) Price() Price          { return p.price }
func (p *Product) Stock() int            { return p.stock }
func (p *Product) Size() string          { return p.size }
func (p *Product) ImageURL() string      { return p.imageURL }
func (p *Product) CreatedAt() time.Time  { return p.createdAt }
