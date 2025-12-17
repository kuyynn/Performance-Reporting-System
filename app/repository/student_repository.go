package repository

import (
	"context"
	"database/sql"
	"uas/app/model"

	"time"
)

type StudentRepository interface {
	// STUDENT PROFILE
	Create(ctx context.Context, s model.StudentCreate) (string, error)
	UpdateProfile(ctx context.Context, s model.StudentCreate) error
	DeleteByUserID(ctx context.Context, userID int64) (bool, error)
	GetStudentID(ctx context.Context, userID int64) (string, error)
	GetStudentsByAdvisor(ctx context.Context, advisorID int64) ([]string, error)
	GetStudentAchievements(ctx context.Context, studentID string) ([]map[string]interface{}, error)

	// ADMIN
	GetAll(ctx context.Context) ([]map[string]interface{}, error)
	GetByID(ctx context.Context, studentID string) (map[string]interface{}, error)

	// ADDITIONAL METHODS
	LecturerExists(ctx context.Context, lecturerID int64) (bool, error)
	AssignAdvisor(ctx context.Context, studentID string, lecturerID int64) error
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

// ADMIN: GET ALL STUDENTS
func (r *StudentRepositoryImpl) GetAll(ctx context.Context) ([]map[string]interface{}, error) {
	query := `
		SELECT
			s.id,
			u.username,
			u.full_name,
			u.email,
			s.points,
			l.id AS lecturer_id,
			u2.full_name AS lecturer_name
		FROM students s
		JOIN users u ON u.id = s.user_id
		LEFT JOIN lecturers l ON l.id = s.advisor_id
		LEFT JOIN users u2 ON u2.id = l.user_id
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
			id           string
			username     string
			fullName     string
			email        string
			points       float64
			lecturerID   sql.NullInt64
			lecturerName sql.NullString
		)

		if err := rows.Scan(
			&id,
			&username,
			&fullName,
			&email,
			&points,
			&lecturerID,
			&lecturerName,
		); err != nil {
			return nil, err
		}

		row := map[string]interface{}{
			"id":       id,
			"username": username,
			"name":     fullName,
			"email":    email,
			"points":   points,
			"advisor":  nil,
		}

		if lecturerID.Valid {
			row["advisor"] = map[string]interface{}{
				"id":   lecturerID.Int64,
				"name": lecturerName.String,
			}
		}

		results = append(results, row)
	}
	return results, nil
}

// ADMIN: GET STUDENT BY ID
func (r *StudentRepositoryImpl) GetByID(ctx context.Context, studentID string) (map[string]interface{}, error) {
	query := `
		SELECT
			s.id,
			u.username,
			u.full_name,
			u.email,
			s.points,
			l.id AS lecturer_id,
			u2.full_name AS lecturer_name
		FROM students s
		JOIN users u ON u.id = s.user_id
		LEFT JOIN lecturers l ON l.id = s.advisor_id
		LEFT JOIN users u2 ON u2.id = l.user_id
		WHERE s.id = $1
		LIMIT 1
	`

	var (
		id           string
		username     string
		fullName     string
		email        string
		points       float64
		lecturerID   sql.NullInt64
		lecturerName sql.NullString
	)

	err := r.DB.QueryRowContext(ctx, query, studentID).Scan(
		&id,
		&username,
		&fullName,
		&email,
		&points,
		&lecturerID,
		&lecturerName,
	)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"id":       id,
		"username": username,
		"name":     fullName,
		"email":    email,
		"points":   points,
		"advisor":  nil,
	}

	if lecturerID.Valid {
		result["advisor"] = map[string]interface{}{
			"id":   lecturerID.Int64,
			"name": lecturerName.String,
		}
	}
	return result, nil
}

// Cek lecturer exists
func (r *StudentRepositoryImpl) LecturerExists(ctx context.Context, lecturerID int64) (bool, error) {
	query := `SELECT COUNT(*) FROM lecturers WHERE id = $1`
	var count int
	err := r.DB.QueryRowContext(ctx, query, lecturerID).Scan(&count)
	return count > 0, err
}

// Assign / update advisor
// ADMIN: PUT /students/:id/advisor
func (r *StudentRepositoryImpl) AssignAdvisor(ctx context.Context, studentID string, lecturerID int64) error {
	query := `
		UPDATE students
		SET advisor_id = $1
		WHERE id = $2
	`
	_, err := r.DB.ExecContext(ctx, query, lecturerID, studentID)
	return err
}

// ADMIN: GET STUDENT ACHIEVEMENTS
func (r *StudentRepositoryImpl) GetStudentAchievements(
	ctx context.Context,
	studentID string,
) ([]map[string]interface{}, error) {
	query := `
		SELECT
			mongo_achievement_id,
			status,
			submitted_at,
			verified_at,
			verified_by,
			rejection_note,
			created_at,
			updated_at
		FROM achievement_references
		WHERE student_uuid = $1
		  AND is_deleted = FALSE
		ORDER BY created_at DESC
	`
	rows, err := r.DB.QueryContext(ctx, query, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []map[string]interface{}
	for rows.Next() {
		var (
			mongoID       string
			status        string
			submittedAt   sql.NullTime
			verifiedAt    sql.NullTime
			verifiedBy    sql.NullInt64
			rejectionNote sql.NullString
			createdAt     time.Time
			updatedAt     time.Time
		)
		if err := rows.Scan(
			&mongoID,
			&status,
			&submittedAt,
			&verifiedAt,
			&verifiedBy,
			&rejectionNote,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, err
		}
		row := map[string]interface{}{
			"mongo_id":   mongoID,
			"status":     status,
			"created_at": createdAt,
			"updated_at": updatedAt,
		}
		if submittedAt.Valid {
			row["submitted_at"] = submittedAt.Time
		}
		if verifiedAt.Valid {
			row["verified_at"] = verifiedAt.Time
		}
		if verifiedBy.Valid {
			row["verified_by"] = verifiedBy.Int64
		}
		if rejectionNote.Valid {
			row["rejection_note"] = rejectionNote.String
		}
		results = append(results, row)
	}
	return results, nil
}
