package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"
	"github.com/lib/pq"
)

type AchievementRepository struct {
	DB *sql.DB
}

func NewAchievementRepository(db *sql.DB) *AchievementRepository {
	return &AchievementRepository{DB: db}
}

// Ambil student UUID berdasarkan user_id
func (r *AchievementRepository) GetStudentID(ctx context.Context, userID int64) (string, error) {
	var studentID string
	query := `
        SELECT id 
        FROM students 
        WHERE user_id = $1
    `
	err := r.DB.QueryRowContext(ctx, query, userID).Scan(&studentID)
	return studentID, err
}

// Insert reference ke PostgreSQL setelah Mongo success
func (r *AchievementRepository) InsertReference(
	ctx context.Context,
	studentID string,
	mongoID string,
) error {
	query := `
        INSERT INTO achievement_references (
            id, student_id, mongo_achievement_id, status, created_at, updated_at
        ) VALUES (
            gen_random_uuid(), $1, $2, 'draft', NOW(), NOW()
        )
    `
	_, err := r.DB.ExecContext(ctx, query, studentID, mongoID)
	return err
}

// Update status dari draft menjadi submitted
func (r *AchievementRepository) Submit(ctx context.Context, achievementID string, studentID string) error {
	query := `
        UPDATE achievement_references
        SET status = 'submitted',
            submitted_at = NOW(),
            updated_at = NOW()
        WHERE mongo_achievement_id = $1
		AND student_id = $2
		AND status = 'draft'
    `
	result, err := r.DB.ExecContext(ctx, query, achievementID, studentID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("cannot submit: not found or not in draft status")
	}
	return nil
}

// Ambil semua achievement milik student (di PostgreSQL)
func (r *AchievementRepository) GetByStudentID(ctx context.Context, studentID string) ([]map[string]interface{}, error) {
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
        WHERE student_id = $1
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
			submittedAt   *time.Time
			verifiedAt    *time.Time
			verifiedBy    *int64
			rejectionNote *string
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
			"mongo_id":       mongoID,
			"status":         status,
			"submitted_at":   submittedAt,
			"verified_at":    verifiedAt,
			"verified_by":    verifiedBy,
			"rejection_note": rejectionNote,
			"created_at":     createdAt,
			"updated_at":     updatedAt,
		}
		results = append(results, row)
	}
	return results, nil
}

// Ambil semua student_id yang dibimbing dosen tertentu
func (r *AchievementRepository) GetStudentsByAdvisor(ctx context.Context, advisorID int64) ([]string, error) {
    query := `
        SELECT id 
        FROM students
        WHERE advisor_id = $1
    `
    rows, err := r.DB.QueryContext(ctx, query, advisorID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var studentIDs []string
    for rows.Next() {
        var sid string
        if err := rows.Scan(&sid); err != nil {
            return nil, err
        }
        studentIDs = append(studentIDs, sid)
    }
    return studentIDs, nil
}

// Ambil semua achievement_references untuk banyak student sekaligus
func (r *AchievementRepository) GetReferencesByStudentList(ctx context.Context, studentIDs []string) ([]map[string]interface{}, error) {
    if len(studentIDs) == 0 {
        return []map[string]interface{}{}, nil
    }
    query := `
        SELECT 
            student_id,
            mongo_achievement_id,
            status,
            submitted_at,
            verified_at,
            verified_by,
            rejection_note,
            created_at,
            updated_at
        FROM achievement_references
        WHERE student_id = ANY($1)
        ORDER BY created_at DESC
    `
    rows, err := r.DB.QueryContext(ctx, query, pq.Array(studentIDs))
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var results []map[string]interface{}
    for rows.Next() {
        var (
            studentID     string
            mongoID       string
            status        string
            submittedAt   *time.Time
            verifiedAt    *time.Time
            verifiedBy    *int64
            rejectionNote *string
            createdAt     time.Time
            updatedAt     time.Time
        )
        if err := rows.Scan(
            &studentID,
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
        results = append(results, map[string]interface{}{
            "student_id":     studentID,
            "mongo_id":       mongoID,
            "status":         status,
            "submitted_at":   submittedAt,
            "verified_at":    verifiedAt,
            "verified_by":    verifiedBy,
            "rejection_note": rejectionNote,
            "created_at":     createdAt,
            "updated_at":     updatedAt,
        })
    }
    return results, nil
}

// Ambil lecturer_id berdasarkan user_id dosen
func (r *AchievementRepository) GetLecturerID(ctx context.Context, userID int64) (int64, error) {
    query := `
        SELECT id
        FROM lecturers
        WHERE user_id = $1
        LIMIT 1
    `
    var lecturerID int64
    err := r.DB.QueryRowContext(ctx, query, userID).Scan(&lecturerID)
    if err != nil {
        return 0, err
    }
    return lecturerID, nil
}
