package repository

import (
	"context"
	"database/sql"
	"errors"
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
