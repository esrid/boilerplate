// PAckage model defition of struct
package models

import (
	"context"
	"time"
)

type User struct {
	ID                  string     `json:"id"`
	Email               string     `json:"email"`
	PasswordHash        string     `json:"-"`
	FailedLoginAttempts int        `json:"-"`
	LockedAt            *time.Time `json:"-"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type LoginPayload struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required"`
}

type NewUser struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=8"`
	Repeat   string `validate:"required,eqfield=Password"`
}

type Session struct {
	ID        string
	UserID    string
	CsrfToken string
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
}

type contextKey string

const userContextKey = contextKey("user")

func UserFromContext(ctx context.Context) *User {
	user, ok := ctx.Value(userContextKey).(*User)
	if !ok {
		return nil
	}
	return user
}

