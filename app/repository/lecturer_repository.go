package repository

import (
	"context"
	"database/sql"
	"uas/app/model"
)

type LecturerRepository interface {
	// LECTURER PROFILE
	Create(ctx context.Context, l model.LecturerCreate) (int64, error)
	Update(ctx context.Context, l model.LecturerCreate) error
	DeleteByUserID(ctx context.Context, userID int64) error
	GetLecturerID(ctx context.Context, userID int64) (int64, error)

	// ADMIN
	GetAll(ctx context.Context) ([]map[string]interface{}, error)
	GetAdvisees(ctx context.Context, lecturerID int64) ([]map[string]interface{}, error)
	GetLecturerIDByUserID(ctx context.Context, userID int64) (int64, error)
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

// DELETE LECTURER PROFILE BY USER ID
func (r *LecturerRepositoryImpl) DeleteByUserID(ctx context.Context, userID int64) error {
    query := `DELETE FROM lecturers WHERE user_id = $1`
    _, err := r.DB.ExecContext(ctx, query, userID)
    return err
}

// ADMIN: GET ALL LECTURERS
func (r *LecturerRepositoryImpl) GetAll(ctx context.Context) ([]map[string]interface{}, error) {
	query := `
		SELECT
			l.id,
			u.full_name,
			u.email,
			COUNT(s.id) AS total_students
		FROM lecturers l
		JOIN users u ON u.id = l.user_id
		LEFT JOIN students s ON s.advisor_id = l.id
		GROUP BY l.id, u.full_name, u.email
		ORDER BY u.full_name
	`
	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []map[string]interface{}
	for rows.Next() {
		var (
			id            int64
			name          string
			email         string
			totalStudents int
		)
		if err := rows.Scan(&id, &name, &email, &totalStudents); err != nil {
			return nil, err
		}
		results = append(results, map[string]interface{}{
			"id":             id,
			"name":           name,
			"email":          email,
			"total_students": totalStudents,
		})
	}
	return results, nil
}

// DOSEN WALI: GET ADVISEES
func (r *LecturerRepositoryImpl) GetAdvisees(ctx context.Context, lecturerID int64) ([]map[string]interface{}, error) {
	query := `
		SELECT
			s.id,
			u.username,
			u.full_name,
			s.points
		FROM students s
		JOIN users u ON u.id = s.user_id
		WHERE s.advisor_id = $1
		ORDER BY u.full_name
	`
	rows, err := r.DB.QueryContext(ctx, query, lecturerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []map[string]interface{}
	for rows.Next() {
		var (
			id       string
			username string
			fullName string
			points   float64
		)
		if err := rows.Scan(&id, &username, &fullName, &points); err != nil {
			return nil, err
		}
		results = append(results, map[string]interface{}{
			"id":       id,
			"username": username,
			"name":     fullName,
			"points":   points,
		})
	}
	return results, nil
}

// helper: get lecturer_id by user_id
func (r *LecturerRepositoryImpl) GetLecturerIDByUserID(ctx context.Context, userID int64) (int64, error) {
	query := `SELECT id FROM lecturers WHERE user_id = $1`
	var id int64
	err := r.DB.QueryRowContext(ctx, query, userID).Scan(&id)
	return id, err
}