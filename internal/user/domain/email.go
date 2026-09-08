package domain

import (
	"errors"
	"regexp"
	"strings"
)

var (
	emailRegex     = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	ErrInvalidEmail = errors.New("email: invalid format")
)

type Email struct{ value string }

func NewEmail(raw string) (Email, error) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	if !emailRegex.MatchString(normalized) {
		return Email{}, ErrInvalidEmail
	}
	return Email{value: normalized}, nil
}

func (e Email) String() string { return e.value }
