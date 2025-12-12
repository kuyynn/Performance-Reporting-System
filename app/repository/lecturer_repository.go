package repository

import (
	"context"
	"database/sql"
	"uas/app/model"
)

type LecturerRepository interface {
	Create(ctx context.Context, l model.LecturerCreate) (int64, error)
	Update(ctx context.Context, l model.LecturerCreate) error
	DeleteByUserID(ctx context.Context, userID int64) error
	GetLecturerID(ctx context.Context, userID int64) (int64, error)
}

type LecturerRepositoryImpl struct {
	DB *sql.DB
}

func NewLecturerRepository(db *sql.DB) LecturerRepository {
	return &LecturerRepositoryImpl{DB: db}
}

// CREATE LECTURER PROFILE
func (r *LecturerRepositoryImpl) Create(ctx context.Context, l model.LecturerCreate) (int64, error) {
	query := `
		INSERT INTO lecturers (user_id, nip, department)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	var id int64
	err := r.DB.QueryRowContext(ctx, query,
		l.UserID,
		l.NIP,
		l.Department,
	).Scan(&id)
	return id, err
}

// GET lecturer_id BY user_id
func (r *LecturerRepositoryImpl) GetLecturerID(ctx context.Context, userID int64) (int64, error) {
	query := `SELECT id FROM lecturers WHERE user_id = $1`
	var id int64
	err := r.DB.QueryRowContext(ctx, query, userID).Scan(&id)
	return id, err
}

// UPDATE LECTURER PROFILE
func (r *LecturerRepositoryImpl) Update(ctx context.Context, l model.LecturerCreate) error {
	query := `
        UPDATE lecturers
        SET nip=$1, department=$2
        WHERE user_id=$3
    `
	_, err := r.DB.ExecContext(ctx, query,
		l.NIP,
		l.Department,
		l.UserID,
	)
	return err
}

func (r *LecturerRepositoryImpl) DeleteByUserID(ctx context.Context, userID int64) error {
    query := `DELETE FROM lecturers WHERE user_id = $1`
    _, err := r.DB.ExecContext(ctx, query, userID)
    return err
}
