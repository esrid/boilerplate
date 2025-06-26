package users

import (
	"context"
	"fmt"

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

func (r *repository) getUserByMail(ctx context.Context, mail string) (*models.User, error) {
	query := fmt.Sprintf("SELECT %s FROM users WHERE email = $1", shared.UserAttribute)

	var user models.User
	var googleID pgtype.Text
	var passwordHash string

	err := r.pool.QueryRow(ctx, query, mail).Scan(
		&user.ID,
		&user.Email,
		&passwordHash,
		&googleID,
		&user.Verify,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	user.GoogleID = googleID.String
	user.Password = passwordHash
	return &user, nil
}

func (r *repository) getUserByID(ctx context.Context, userID string) (*models.User, error) {
	query := fmt.Sprintf("SELECT %s FROM users WHERE id = $1", shared.UserAttribute)

	var user models.User
	var googleID pgtype.Text
	var passwordHash string

	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Email,
		&passwordHash,
		&googleID,
		&user.Verify,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	user.GoogleID = googleID.String
	user.Password = passwordHash
	return &user, nil
}

func (r *repository) update(ctx context.Context, u models.User) (*models.User, error) {
	query := fmt.Sprintf(
		"UPDATE users SET email = $1, password_hash = $2, updated_at = NOW() RETURNING %s",
		shared.UserAttribute,
	)

	var user models.User
	var googleID pgtype.Text
	var passwordHash string

	err := r.pool.QueryRow(ctx, query, u.Email, u.Password).Scan(
		&user.ID,
		&user.Email,
		&passwordHash,
		&googleID,
		&user.Verify,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	user.GoogleID = googleID.String
	user.Password = passwordHash
	return &user, nil
}

func (r *repository) delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

func (r *repository) updateVerify(ctx context.Context, id string, verify bool) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET verify = $1 WHERE id = $2`, verify, id)
	if err != nil {
		return fmt.Errorf("failed to update user verification status: %w", err)
	}
	return nil
}
