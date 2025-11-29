package repository

import (
	"context"
	"database/sql"
	// "encoding/json"
	// "fmt"
	"uas/app/model"
)

// Interface untuk UserRepository
type UserRepository interface {
	Create(ctx context.Context, user model.UserCreateRequest) (int64, error)
	Update(ctx context.Context, user model.UserUpdateRequest) error
	Delete(ctx context.Context, userID int64) error
	FindById(ctx context.Context, userID int64) (*model.UserResponse, error)
	FindAll(ctx context.Context) (*[]model.UserResponse, error)
	FindByUsernameOrEmail(usernameOrEmail string) (*model.User, error)
	GetPermissionsByUserID(userID int64) ([]model.Permission, error)
	Logout()
}

// Implementasi UserRepository
type UserRepositoryy struct {
	DB *sql.DB
}

// CRUD
func (r *UserRepositoryy) Create(ctx context.Context, user model.UserCreateRequest) (int64, error) {
	sql := `INSERT INTO users(username, full_name, email, password_hash, role_id) VALUES (?, ?, ?, ?, ?)`
	res, err := r.DB.ExecContext(ctx, sql, user.Username, user.FullName, user.Email, user.Password)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *UserRepositoryy) Update(ctx context.Context, user model.UserUpdateRequest) error {
	sql := `UPDATE users SET username=?, full_name=?, email=?, password_hash=?, role_id=? WHERE id=?`
	_, err := r.DB.ExecContext(ctx, sql, user.Username, user.FullName, user.Email, user.ID)
	return err
}

func (r *UserRepositoryy) Delete(ctx context.Context, userID int64) error {
	sql := `DELETE FROM users WHERE id=?`
	_, err := r.DB.ExecContext(ctx, sql, userID)
	return err
}

// Find
func (r *UserRepositoryy) FindAll(ctx context.Context) (*[]model.UserResponse, error) {
	sql := `SELECT id, username, full_name, role FROM users`
	rows, err := r.DB.QueryContext(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.UserResponse
	for rows.Next() {
		var u model.UserResponse
		if err := rows.Scan(&u.ID, &u.Username, &u.FullName, &u.Role); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return &users, nil
}

func (r *UserRepositoryy) FindById(ctx context.Context, id int64) (*model.UserResponse, error) {
	sql := `SELECT id, username, full_name, role FROM users WHERE id=?`
	row := r.DB.QueryRowContext(ctx, sql, id)

	var u model.UserResponse
	if err := row.Scan(&u.ID, &u.Username, &u.FullName, &u.Role); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepositoryy) FindByUsernameOrEmail(usernameOrEmail string) (*model.User, error) {
	sql := `SELECT u.id, u.username, u.full_name, u.email, u.password_hash, r.name
			FROM users u
			INNER JOIN roles r ON u.role_id = r.id
			WHERE u.username=? OR u.email=?`
	row := r.DB.QueryRow(sql, usernameOrEmail, usernameOrEmail)

	var u model.User
	if err := row.Scan(&u.ID, &u.Username, &u.FullName, &u.Email, &u.PasswordHash, &u.RoleID, &u.Role); err != nil {
		return nil, err
	}
	return &u, nil
}

// Permissions 
func (r *UserRepositoryy) GetPermissionsByUserID(userID int64) ([]string, error) {
	query := `
	SELECT p.name
	FROM permissions p
	JOIN role_permissions rp ON rp.permission_id = p.id
	JOIN users u ON u.role_id = rp.role_id
	WHERE u.id = ?`
	rows, err := r.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}
	return perms, nil
}

// Logout 
func (r *UserRepositoryy) Logout() {
	// Kosong dulu, nanti bisa implement session/token revoke
}
