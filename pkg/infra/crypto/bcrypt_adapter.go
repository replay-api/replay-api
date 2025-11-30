package crypto

import (
	"context"

	email_out "github.com/replay-api/replay-api/pkg/domain/email/ports/out"
	"golang.org/x/crypto/bcrypt"
)

type BcryptPasswordHasherAdapter struct {
	cost int
}

func NewBcryptPasswordHasherAdapter(cost int) email_out.PasswordHasher {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}
	return &BcryptPasswordHasherAdapter{cost: cost}
}

func (b *BcryptPasswordHasherAdapter) HashPassword(ctx context.Context, password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), b.cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (b *BcryptPasswordHasherAdapter) ComparePassword(ctx context.Context, hashedPassword string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
