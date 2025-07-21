// Package auth provides authentication-related functionalities.
package auth

import (
	"boilerplate/internal/models"
	"boilerplate/pkg/utils"
	"context"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var ( // TODO: move to a better place
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrUserLocked        = errors.New("user account is locked")
	ErrEmailAlreadyInUse = errors.New("email already in use")
)

const (
	maxLoginAttempts = 5
	lockoutDuration  = 15 * time.Minute
)

type Service interface {
	Register(ctx context.Context, email, password string) (*models.Session, error)
	Login(ctx context.Context, email, password string) (*models.Session, error)
	Logout(ctx context.Context, sessionID string) error
	GetSession(ctx context.Context, sessionID string) (*models.Session, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Register(ctx context.Context, email, password string) (*models.Session, error) {
	_, err := s.repo.FindUserByEmail(ctx, email)
	if err == nil {
		return nil, ErrEmailAlreadyInUse
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.CreateUser(ctx, email, string(passwordHash))
	if err != nil {
		return nil, err
	}
	return s.setSession(ctx, user.ID)
}

func (s *service) Login(ctx context.Context, email, password string) (*models.Session, error) {
	user, err := s.repo.FindUserByEmail(ctx, email)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if user.LockedAt != nil && time.Now().Before(*user.LockedAt) {
		return nil, ErrUserLocked
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		user.FailedLoginAttempts++
		if user.FailedLoginAttempts >= maxLoginAttempts {
			lockedAt := time.Now().Add(lockoutDuration)
			user.LockedAt = &lockedAt
		}
		err := s.repo.UpdateUserLock(ctx, user.ID, user.FailedLoginAttempts, user.LockedAt)
		if err != nil {
			return nil, err
		}
		return nil, ErrInvalidPassword
	}

	user.FailedLoginAttempts = 0
	user.LockedAt = nil
	err = s.repo.UpdateUserLock(ctx, user.ID, user.FailedLoginAttempts, user.LockedAt)
	if err != nil {
		return nil, err
	}

	return s.setSession(ctx, user.ID)
}

func (s *service) Logout(ctx context.Context, sessionID string) error {
	return s.repo.DeleteSession(ctx, sessionID)
}

func (s *service) GetSession(ctx context.Context, sessionID string) (*models.Session, error) {
	return s.repo.FindSessionByID(ctx, sessionID)
}

func (s *service) setSession(ctx context.Context, userID string) (*models.Session, error) {
	_ = s.repo.DeleteSessionByUserID(ctx, userID)

	csrf, token, err := utils.GenerateTokenAndCrf()
	if err != nil {
		return nil, err
	}

	session := models.Session{
		ID:        token,
		UserID:    userID,
		CsrfToken: csrf,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	createdSession, err := s.repo.CreateSession(ctx, session)
	if err != nil {
		return nil, err
	}

	createdSession.Token = token
	return createdSession, nil
}