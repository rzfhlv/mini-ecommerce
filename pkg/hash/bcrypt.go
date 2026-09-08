package hash

import (
	"golang.org/x/crypto/bcrypt"

	userdomain "mini-ecommerce/internal/user/domain"
)

const bcryptCost = 12

type BcryptHasher struct{}

var _ userdomain.PasswordHasher = (*BcryptHasher)(nil)

func NewBcryptHasher() *BcryptHasher { return &BcryptHasher{} }

func (h *BcryptHasher) Hash(plain string) (userdomain.PasswordHash, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil {
		return userdomain.PasswordHash{}, err
	}
	return userdomain.NewPasswordHash(string(hashed))
}

func (h *BcryptHasher) Compare(hash userdomain.PasswordHash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash.String()), []byte(plain)) == nil
}
