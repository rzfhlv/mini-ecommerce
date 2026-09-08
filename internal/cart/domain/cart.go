package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidQuantity  = errors.New("cart: quantity must be greater than zero")
	ErrCartItemNotFound = errors.New("cart: item not found")
)

type CartItem struct {
	productID     uuid.UUID
	productName   string
	priceSnapshot int64 // plain int64 -- tidak ada operasi aritmatika kompleks yang butuh Value Object di sini
	quantity      int
}

func (i CartItem) ProductID() uuid.UUID { return i.productID }
func (i CartItem) ProductName() string  { return i.productName }
func (i CartItem) PriceSnapshot() int64 { return i.priceSnapshot }
func (i CartItem) Quantity() int        { return i.quantity }
func (i CartItem) Subtotal() int64      { return i.priceSnapshot * int64(i.quantity) }

type Cart struct {
	id, userID uuid.UUID
	items      []CartItem
	updatedAt  time.Time
}

func NewCart(userID uuid.UUID) *Cart { return &Cart{id: uuid.New(), userID: userID, items: []CartItem{}, updatedAt: time.Now()} }

func ReconstructCart(id, userID uuid.UUID, items []CartItem, updatedAt time.Time) *Cart {
	return &Cart{id: id, userID: userID, items: items, updatedAt: updatedAt}
}
func ReconstructCartItem(productID uuid.UUID, productName string, price int64, quantity int) CartItem {
	return CartItem{productID: productID, productName: productName, priceSnapshot: price, quantity: quantity}
}

func (c *Cart) AddItem(productID uuid.UUID, productName string, price int64, quantity int) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}
	for i, item := range c.items {
		if item.productID == productID {
			c.items[i].quantity += quantity
			c.items[i].priceSnapshot = price
			c.items[i].productName = productName
			c.updatedAt = time.Now()
			return nil
		}
	}
	c.items = append(c.items, CartItem{productID: productID, productName: productName, priceSnapshot: price, quantity: quantity})
	c.updatedAt = time.Now()
	return nil
}

func (c *Cart) UpdateItemQuantity(productID uuid.UUID, quantity int) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}
	for i, item := range c.items {
		if item.productID == productID {
			c.items[i].quantity = quantity
			c.updatedAt = time.Now()
			return nil
		}
	}
	return ErrCartItemNotFound
}

func (c *Cart) RemoveItem(productID uuid.UUID) error {
	for i, item := range c.items {
		if item.productID == productID {
			c.items = append(c.items[:i], c.items[i+1:]...)
			c.updatedAt = time.Now()
			return nil
		}
	}
	return ErrCartItemNotFound
}

func (c *Cart) Clear() { c.items = []CartItem{}; c.updatedAt = time.Now() }
func (c *Cart) Total() int64 {
	var total int64
	for _, item := range c.items {
		total += item.Subtotal()
	}
	return total
}
func (c *Cart) ID() uuid.UUID     { return c.id }
func (c *Cart) UserID() uuid.UUID { return c.userID }
func (c *Cart) Items() []CartItem { return c.items }
