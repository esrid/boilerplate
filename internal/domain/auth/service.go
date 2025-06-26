package auth

import (
	"context"
	"errors"
	"net/http"
	"time"
	"yourapp/config"
	"yourapp/internal/models"
	"yourapp/internal/shared"
	"yourapp/pkg"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type service struct {
	r *repository
	v *validator.Validate
	k config.APIKEY
}

func NewService(repo *repository, k *config.APIKEY) *service {
	v := validator.New()

	return &service{r: repo, v: v}
}

func (s *service) registerUser(ctx context.Context, r *http.Request, u models.NewUser) (string, error) {
	if err := s.v.Struct(u); err != nil {
		return "", shared.ErrInvalidInputs
	}

	u.Email = pkg.CleanAndLower(u.Email)

	user, err := s.r.GetUserByEmail(ctx, u.Email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}

	if user != nil {
		return "", shared.ErrUserAlreadyExist
	}

	if u.Password != u.Repeat {
		return "", shared.ErrInvalidInputs
	}

	u.Password, _ = pkg.HashPassword(u.Password)

	newUser, err := s.r.createUser(ctx, &u)
	if err != nil {
		return "", err
	}
	token, err := s.tokenAndSession(ctx, newUser.ID, r)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *service) tokenAndSession(ctx context.Context, userID string, r *http.Request) (string, error) {
	token := pkg.GenerateToken()
	csrfCode := pkg.GenerateToken()

	newtoken, err := s.r.createSession(ctx, models.Session{
		UserID:    userID,
		Token:     token,
		CsrfCode:  csrfCode,
		ExpiresAt: time.Now().Add(shared.TwoDays),
		IPAddress: pkg.GetIPAddressBytes(r),
		UserAgent: r.UserAgent(),
	})
	if err != nil {
		return "", err
	}

	return newtoken, nil

}

func (s *service) login(ctx context.Context, r *http.Request, u models.User) (string, error) {
	if err := s.v.Struct(u); err != nil {
		return "", shared.ErrInvalidInputs
	}

	u.Email = pkg.CleanAndLower(u.Email)

	user, err := s.r.GetUserByEmail(ctx, u.Email)
	if err != nil {
		return "", err
	}
	if err2 := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(u.Password)); err != nil {
		return "", err2
	}

	if err3 := s.r.DeleteBySessionToken(ctx, user.ID); err != nil {
		return "", err3
	}

	token, err := s.tokenAndSession(ctx, user.ID, r)
	if err != nil {
		return "", err
	}

	return token, nil
}
