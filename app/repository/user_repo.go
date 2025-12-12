package repository

import (
	"context"
	"database/sql"
	"uas/app/model"
)

type UserRepository interface {
	Create(ctx context.Context, user model.UserCreateRequest) (int64, error)
	CreateRaw(ctx context.Context, username, fullName, email, passHash, roleID string) (int64, error)
	Update(ctx context.Context, user model.UserUpdateRequest) error
	UpdateRaw(ctx context.Context, id int64, username, fullName, email, roleID string) error
	Delete(ctx context.Context, userID int64) error
	GetRoleIDByName(roleName string) (string, error)
	FindById(ctx context.Context, userID int64) (*model.UserResponse, error)
	FindAll(ctx context.Context) (*[]model.UserResponse, error)
	FindByUsernameOrEmail(usernameOrEmail string) (*model.User, error)
	GetPermissionsByUserID(userID int64) ([]string, error)
	Logout()
}

type UserRepositoryImpl struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &UserRepositoryImpl{DB: db}
}

// GET ROLE ID BY NAME
func (r *UserRepositoryImpl) GetRoleIDByName(name string) (string, error) {
	query := `SELECT id FROM roles WHERE name = $1`
	var id string
	err := r.DB.QueryRow(query, name).Scan(&id)
	return id, err
}

// CREATE USER (POSTGRESQL)
func (r *UserRepositoryImpl) Create(ctx context.Context, user model.UserCreateRequest) (int64, error) {
	sqlQuery := `INSERT INTO users(username, full_name, email, password_hash, role_id)
		         VALUES ($1, $2, $3, $4, $5)
				 RETURNING id`
	var id int64
	err := r.DB.QueryRowContext(ctx, sqlQuery,
		user.Username,
		user.FullName,
		user.Email,
		user.Password,
		"default",
	).Scan(&id)
	return id, err
}

// UPDATE USER
func (r *UserRepositoryImpl) Update(ctx context.Context, user model.UserUpdateRequest) error {
	sqlQuery := `UPDATE users SET username=$1, full_name=$2, email=$3 WHERE id=$4`
	_, err := r.DB.ExecContext(ctx, sqlQuery,
		user.Username,
		user.FullName,
		user.Email,
		user.ID,
	)
	return err
}

// DELETE USER
func (r *UserRepositoryImpl) Delete(ctx context.Context, userID int64) error {
	sqlQuery := `DELETE FROM users WHERE id=$1`
	_, err := r.DB.ExecContext(ctx, sqlQuery, userID)
	return err
}

// FIND ALL USERS
func (r *UserRepositoryImpl) FindAll(ctx context.Context) (*[]model.UserResponse, error) {
	sqlQuery := `SELECT id, username, full_name, role_id FROM users`

	rows, err := r.DB.QueryContext(ctx, sqlQuery)
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

// FIND USER BY ID
func (r *UserRepositoryImpl) FindById(ctx context.Context, id int64) (*model.UserResponse, error) {
	sqlQuery := `SELECT id, username, full_name, role_id FROM users WHERE id=$1`

	row := r.DB.QueryRowContext(ctx, sqlQuery, id)

	var u model.UserResponse
	if err := row.Scan(&u.ID, &u.Username, &u.FullName, &u.Role); err != nil {
		return nil, err
	}
	return &u, nil
}

// LOGIN FIND
func (r *UserRepositoryImpl) FindByUsernameOrEmail(usernameOrEmail string) (*model.User, error) {
	sqlQuery := `
        SELECT 
            u.id, u.username, u.full_name, u.email,
            u.password_hash, u.role_id, u.is_active,
            r.name AS role_name
        FROM users u
        JOIN roles r ON r.id = u.role_id
        WHERE u.username=$1 OR u.email=$1
    `
	row := r.DB.QueryRow(sqlQuery, usernameOrEmail)
	var u model.User
	if err := row.Scan(
		&u.ID,
		&u.Username,
		&u.FullName,
		&u.Email,
		&u.PasswordHash,
		&u.RoleID,
		&u.IsActive,
		&u.Role, // ← inilah field Role yang benar!
	); err != nil {
		return nil, err
	}
	return &u, nil
}

// PERMISSIONS
func (r *UserRepositoryImpl) GetPermissionsByUserID(userID int64) ([]string, error) {
	query := `
	SELECT p.name
	FROM permissions p
	JOIN role_permissions rp ON rp.permission_id = p.id
	JOIN users u ON u.role_id = rp.role_id
	WHERE u.id = $1
	`

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

func (r *UserRepositoryImpl) Logout() {}

// CREATE RAW USER (for AdminService)
func (r *UserRepositoryImpl) CreateRaw(ctx context.Context, username, fullName, email, passHash, roleID string) (int64, error) {
	query := `
		INSERT INTO users(username, full_name, email, password_hash, role_id, is_active)
		VALUES ($1,$2,$3,$4,$5,TRUE)
		RETURNING id
	`
	var id int64
	err := r.DB.QueryRowContext(ctx, query,
		username, fullName, email, passHash, roleID,
	).Scan(&id)
	return id, err
}

// UPDATE RAW USER (Admin can update username, fullname, email, role)
func (r *UserRepositoryImpl) UpdateRaw(
    ctx context.Context,
    id int64,
    username, fullName, email, roleID string,
) error {
    query := `
        UPDATE users
        SET username=$1,
            full_name=$2,
            email=$3,
            role_id=$4,
            updated_at = NOW()
        WHERE id=$5
    `
    _, err := r.DB.ExecContext(ctx, query,
        username,
        fullName,
        email,
        roleID,
        id,
    )
    return err
}
