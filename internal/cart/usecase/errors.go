package usecase

import "errors"

var (
	ErrInvalidUserID    = errors.New("usecase: invalid user id")
	ErrInvalidProductID = errors.New("usecase: invalid product id")
)
