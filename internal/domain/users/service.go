package users

import (
	"context"
	"yourapp/internal/models"
	"yourapp/internal/shared"
	"yourapp/pkg"

	"github.com/go-playground/validator"
)

type service struct {
	r *repository
	v *validator.Validate
}

func NewService(repo *repository) *service {
	v := validator.New()
	return &service{r: repo, v: v}
}

func (s *service) update(ctx context.Context, u models.User) (*models.User, error) {
	if err := s.v.Struct(u); err != nil {
		return nil, shared.ErrInvalidInputs
	}
	u.Email = pkg.CleanAndLower(u.Email)

	u.Password, _ = pkg.HashPassword(u.Password)
	user, err := s.r.update(ctx, u)
	if err != nil {
		return nil, err
	}
	return user, nil
}
