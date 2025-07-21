// Package auth provides authentication-related functionalities.
package auth

import (
	"boilerplate/internal/models"
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type Repository interface {
	CreateUser(ctx context.Context, email, passwordHash string) (*models.User, error)
	FindUserByEmail(ctx context.Context, email string) (*models.User, error)
	UpdateUserLock(ctx context.Context, id string, failedLoginAttempts int, lockedAt *time.Time) error

	CreateSession(ctx context.Context, session models.Session) (*models.Session, error)
	FindSessionByID(ctx context.Context, id string) (*models.Session, error)
	DeleteSession(ctx context.Context, id string) error
	DeleteSessionByUserID(ctx context.Context, userID string) error
}

type pgxRepository struct {
	db *pgx.Conn
}

func NewRepository(db *pgx.Conn) Repository {
	return &pgxRepository{db: db}
}

func (r *pgxRepository) CreateUser(ctx context.Context, email, passwordHash string) (*models.User, error) {
	var user models.User
	query := `INSERT INTO users (email, password_hash, failed_login_attempts, locked_at) VALUES ($1, $2, 0, NULL)
		  RETURNING id, email, password_hash, failed_login_attempts, locked_at, created_at, updated_at`
	err := r.db.QueryRow(ctx, query, email, passwordHash).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FailedLoginAttempts, &user.LockedAt, &user.CreatedAt, &user.UpdatedAt)
	return &user, err
}

func (r *pgxRepository) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	query := `SELECT id, email, password_hash, failed_login_attempts, locked_at, created_at, updated_at FROM users WHERE email = $1`
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FailedLoginAttempts,
		&user.LockedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	return &user, err
}

func (r *pgxRepository) UpdateUserLock(ctx context.Context, id string, failedLoginAttempts int, lockedAt *time.Time) error {
	query := `UPDATE users SET failed_login_attempts = $1, locked_at = $2, updated_at = $3 WHERE id = $4`
	_, err := r.db.Exec(ctx, query, failedLoginAttempts, lockedAt, time.Now(), id)
	return err
}

func (r *pgxRepository) CreateSession(ctx context.Context, session models.Session) (*models.Session, error) {
	var ses models.Session
	query := `INSERT INTO sessions (user_id, token, csrf_token, expires_at) VALUES ($1, $2, $3,$4)
				 RETURNING id, user_id, csrf_token, expires_at, created_at`
	err := r.db.QueryRow(ctx, query, session.UserID, session.Token, session.CsrfToken, time.Now().Add(24*time.Hour)).Scan(&ses.ID, &ses.UserID, &ses.CsrfToken, &ses.ExpiresAt, &ses.CreatedAt)
	return &ses, err
}

func (r *pgxRepository) FindSessionByID(ctx context.Context, id string) (*models.Session, error) {
	var session models.Session
	query := `SELECT id, user_id, csrf_token, expires_at, created_at FROM sessions WHERE id = $1`
	err := r.db.QueryRow(ctx, query, id).Scan(&session.ID, &session.UserID, &session.CsrfToken, &session.ExpiresAt, &session.CreatedAt)
	return &session, err
}

func (r *pgxRepository) DeleteSession(ctx context.Context, id string) error {
	query := `DELETE FROM sessions WHERE user_id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *pgxRepository) DeleteSessionByUserID(ctx context.Context, userID string) error {
	query := `DELETE FROM sessions WHERE user_id = $1`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}
