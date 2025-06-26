package auth

import (
	"context"
	"fmt"
	"time"
	"yourapp/internal/models"
	"yourapp/internal/shared"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *repository {
	return &repository{pool: pool}
}

type Authstore interface {
	GetUserBySessionID(ctx context.Context, sid string) (*models.User, error)
	GetBySessionToken(ctx context.Context, cookieHash string) (*models.Session, error)
	UpdateExpiry(ctx context.Context, token string, expiresAt time.Time) error
	createSession(ctx context.Context, s models.Session) (string, error)
	DeleteBySessionToken(ctx context.Context, token string) error
	DeletePreviousSession(ctx context.Context, user_id string) error
	createUser(ctx context.Context, u *models.User) (*models.User, error)
	createUserWithGoogle(ctx context.Context, u *models.User) (*models.User, error)
	setOauthToTrue(ctx context.Context, email string) error
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
}

func (r *repository) createSession(ctx context.Context, s models.Session) (string, error) {
	var cookieHash string
	query := fmt.Sprintf(`INSERT INTO sessions (%s) VALUES ($1, $2, $3, NOW(), $4, $5, $6) RETURNING token`, shared.SessionAttributes)
	if err := r.pool.QueryRow(ctx, query, s.UserID, s.Token, s.CsrfCode, s.ExpiresAt.UTC(), s.IPAddress.String(), s.UserAgent).Scan(&cookieHash); err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	return cookieHash, nil
}

func (r *repository) GetBySessionToken(ctx context.Context, token string) (*models.Session, error) {
	var s models.Session
	query := `SELECT user_id, token, csrf_code, expires_at FROM sessions WHERE token = $1`
	if err := r.pool.QueryRow(ctx, query, token).Scan(&s.UserID, &s.Token, &s.CsrfCode, &s.ExpiresAt); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *repository) DeleteBySessionToken(ctx context.Context, token string) error {
	if _, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE token = $1`, token); err != nil {
		return fmt.Errorf("failed to delete session by token: %w", err)
	}
	return nil
}

func (r *repository) DeletePreviousSession(ctx context.Context, user_id string) error {
	if _, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE user_id = $1`, user_id); err != nil {
		return fmt.Errorf("failed to delete previous sessions for user %s: %w", user_id, err)
	}
	return nil
}

func (r *repository) UpdateExpiry(ctx context.Context, token string, expiresAt time.Time) error {
	if _, err := r.pool.Exec(ctx, `UPDATE sessions SET expires_at = $1 WHERE token = $2`, expiresAt, token); err != nil {
		return fmt.Errorf("failed to update session expiry: %w", err)
	}
	return nil
}

func (r *repository) createUser(ctx context.Context, u *models.NewUser) (*models.User, error) {
	user := &models.User{}
	var googleID pgtype.Text
	var passwordHash string
	query := fmt.Sprintf(`INSERT INTO users (email, password_hash, created_at, updated_at) VALUES ($1, $2, NOW(), NOW()) RETURNING %s`, shared.UserAttribute)
	if err := r.pool.QueryRow(ctx, query, u.Email, u.Password).Scan(&user.ID, &user.Email, &passwordHash, &googleID, &user.Oauth, &user.Verify, &user.Role,
		&user.CreatedAt, &user.UpdatedAt); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	user.GoogleID = googleID.String
	user.Password = ""
	return user, nil
}

func (r *repository) createUserWithGoogle(ctx context.Context, u *models.User) (*models.User, error) {
	user := &models.User{}
	var googleID pgtype.Text
	var passwordHash string

	query := fmt.Sprintf(`INSERT INTO users (email, google_id, oauth, created_at, updated_at) VALUES ($1, $2, $3, NOW(), NOW()) RETURNING %s`, shared.UserAttribute)
	if err := r.pool.QueryRow(ctx, query, u.Email, u.GoogleID, u.Oauth).Scan(&user.ID, &user.Email, &passwordHash, &googleID, &user.Oauth, &user.Verify, &user.Role,
		&user.CreatedAt, &user.UpdatedAt); err != nil {
		return nil, fmt.Errorf("failed to create user with Google: %w", err)
	}

	user.GoogleID = googleID.String
	user.Password = ""
	return user, nil
}

func (r *repository) setOauthToTrue(ctx context.Context, email string) error {
	if _, err := r.pool.Exec(ctx, "UPDATE users SET oauth = TRUE WHERE email = $1", email); err != nil {
		return fmt.Errorf("failed to set oauth to true for email %s: %w", email, err)
	}
	return nil
}

func (r *repository) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	user := &models.User{}
	var googleID pgtype.Text
	var passwordHash string

	query := fmt.Sprintf(`SELECT %s FROM users WHERE id = $1`, shared.UserAttribute)
	if err := r.pool.QueryRow(ctx, query, id).Scan(&user.ID, &user.Email, &passwordHash, &googleID, &user.Oauth, &user.Verify, &user.Role, &user.CreatedAt, &user.UpdatedAt); err != nil {
		return nil, err
	}

	user.GoogleID = googleID.String
	user.Password = ""
	return user, nil
}

func (r *repository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	var googleID pgtype.Text
	var passwordHash string

	query := fmt.Sprintf(`SELECT %s FROM users WHERE email = $1`, shared.UserAttribute)
	if err := r.pool.QueryRow(ctx, query, email).Scan(&user.ID, &user.Email, &passwordHash, &googleID, &user.Oauth, &user.Verify, &user.Role, &user.CreatedAt,
		&user.UpdatedAt); err != nil {
		return nil, err
	}

	user.GoogleID = googleID.String
	user.Password = passwordHash
	return user, nil
}

func (r *repository) GetUserBySessionID(ctx context.Context, token string) (*models.User, error) {
	u := &models.User{}
	if err := r.pool.QueryRow(ctx, `SELECT u.id, u.email, u.updated_at FROM users u INNER JOIN sessions s ON u.id = s.user_id WHERE s.token = $1`, token).Scan(&u.ID, &u.Email, &u.UpdatedAt); err != nil {
		return nil, err
	}
	return u, nil
}
