package repository

import (
	"context"
	"database/sql"
	"uas/app/model"
)

type StudentRepository interface {
	Create(ctx context.Context, s model.StudentCreate) (string, error)
	UpdateProfile(ctx context.Context, s model.StudentCreate) error
	DeleteByUserID(ctx context.Context, userID int64) (bool, error)
	GetStudentID(ctx context.Context, userID int64) (string, error)
	GetStudentsByAdvisor(ctx context.Context, advisorID int64) ([]string, error)
}

type StudentRepositoryImpl struct {
	DB *sql.DB
}

func NewStudentRepository(db *sql.DB) StudentRepository {
	return &StudentRepositoryImpl{DB: db}
}

// CREATE STUDENT PROFILE
func (r *StudentRepositoryImpl) Create(ctx context.Context, s model.StudentCreate) (string, error) {
	query := `
		INSERT INTO students (user_id, program_study, academic_year, advisor_id, nim)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	var id string
	err := r.DB.QueryRowContext(ctx, query,
		s.UserID,
		s.ProgramStudy,
		s.AcademicYear,
		s.AdvisorID,
		s.Nim,
	).Scan(&id)
	return id, err
}

// GET STUDENT UUID BY USER ID
func (r *StudentRepositoryImpl) GetStudentID(ctx context.Context, userID int64) (string, error) {
	query := `SELECT id FROM students WHERE user_id = $1`
	var id string
	err := r.DB.QueryRowContext(ctx, query, userID).Scan(&id)
	return id, err
}

// GET ALL STUDENTS UNDER A SPECIFIC ADVISOR
func (r *StudentRepositoryImpl) GetStudentsByAdvisor(ctx context.Context, advisorID int64) ([]string, error) {
	query := `SELECT id FROM students WHERE advisor_id = $1`
	rows, err := r.DB.QueryContext(ctx, query, advisorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var sid string
		if err := rows.Scan(&sid); err != nil {
			return nil, err
		}
		ids = append(ids, sid)
	}
	return ids, nil
}

// UPDATE STUDENT PROFILE
func (r *StudentRepositoryImpl) UpdateProfile(ctx context.Context, s model.StudentCreate) error {
	query := `
        UPDATE students
        SET program_study=$1, academic_year=$2, advisor_id=$3
        WHERE user_id=$4
    `
	_, err := r.DB.ExecContext(ctx, query,
		s.ProgramStudy,
		s.AcademicYear,
		s.AdvisorID,
		s.UserID,
	)
	return err
}

func (r *StudentRepositoryImpl) DeleteByUserID(ctx context.Context, userID int64) (bool, error) {
    query := `DELETE FROM students WHERE user_id = $1`
    _, err := r.DB.ExecContext(ctx, query, userID)
    return err == nil, err
}
