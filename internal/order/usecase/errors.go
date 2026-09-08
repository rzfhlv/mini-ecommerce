package usecase

import "errors"

var (
	ErrInvalidUserID    = errors.New("usecase: invalid user id")
	ErrInvalidProductID = errors.New("usecase: invalid product id")
	ErrInvalidPromoID   = errors.New("usecase: invalid promo id")
	ErrOrderNotFound    = errors.New("usecase: order not found")
)
