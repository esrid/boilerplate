package app

import "github.com/go-playground/validator"

type service struct {
	r *repository
	v *validator.Validate
}

func NewService(repo *repository) *service {
	v := validator.New()
	return &service{r: repo, v: v}
}
