package repository

import (
	"context"
	"database/sql"
	"errors"
)

type RoleRepository struct {
	DB *sql.DB
}

func NewRoleRepository(db *sql.DB) *RoleRepository {
	return &RoleRepository{DB: db}
}

func (r *RoleRepository) GetRoleIDByName(ctx context.Context, name string) (string, error) {
	query := `SELECT id FROM roles WHERE name=$1`

	var id string
	err := r.DB.QueryRowContext(ctx, query, name).Scan(&id)
	if err != nil {
		return "", errors.New("role_not_found")
	}
	return id, nil
}
