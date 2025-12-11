package repository

import (
	"context"
	"database/sql"
	"time"
	"uas/app/model"
)

type RefreshTokenRepository interface {
	Save(ctx context.Context, userID int64, token string, expiresAt time.Time) error
	Get(ctx context.Context, token string) (*model.RefreshToken, error)
	DeleteByUserID(ctx context.Context, userID int64) error
}

type RefreshTokenRepositoryImpl struct {
	DB *sql.DB
}

func NewRefreshTokenRepository(db *sql.DB) RefreshTokenRepository {
	return &RefreshTokenRepositoryImpl{DB: db}
}

func (r *RefreshTokenRepositoryImpl) Save(ctx context.Context, userID int64, token string, expiresAt time.Time) error {
	query := `
        INSERT INTO refresh_tokens (user_id, token, expires_at)
        VALUES ($1, $2, $3)
    `
	_, err := r.DB.ExecContext(ctx, query, userID, token, expiresAt)
	return err
}

func (r *RefreshTokenRepositoryImpl) Get(ctx context.Context, token string) (*model.RefreshToken, error) {
	query := `SELECT id, user_id, token, expires_at, created_at FROM refresh_tokens WHERE token=$1`

	row := r.DB.QueryRowContext(ctx, query, token)

	var rt model.RefreshToken
	err := row.Scan(&rt.ID, &rt.UserID, &rt.Token, &rt.ExpiresAt, &rt.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *RefreshTokenRepositoryImpl) DeleteByUserID(ctx context.Context, userID int64) error {
	_, err := r.DB.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE user_id=$1`, userID)
	return err
}
