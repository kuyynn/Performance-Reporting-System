package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
)

type AchievementRepository struct {
	DB *sql.DB
}

type AchievementAdminFilter struct {
	Status    string
	StudentID string
	Sort      string
	Order     string
	Limit     int
	Offset    int
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
            id, student_uuid, mongo_achievement_id, status, created_at, updated_at
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
		AND student_uuid = $2
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
func (r *AchievementRepository) GetByStudentID(ctx context.Context, studentID string, includeDeleted bool) ([]map[string]interface{}, error) {
	query := `
        SELECT mongo_achievement_id, status, submitted_at, verified_at, rejection_note
        FROM achievement_references
        WHERE student_uuid = $1
        AND ($2 OR is_deleted = false);
    `
	rows, err := r.DB.QueryContext(ctx, query, studentID, includeDeleted)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []map[string]interface{}
	for rows.Next() {
		var mongoID, status string
		var submittedAt, verifiedAt, rejectionNote sql.NullString
		if err := rows.Scan(&mongoID, &status, &submittedAt, &verifiedAt, &rejectionNote); err != nil {
			return nil, err
		}
		result = append(result, map[string]interface{}{
			"mongo_id":       mongoID,
			"status":         status,
			"submitted_at":   submittedAt.String,
			"verified_at":    verifiedAt.String,
			"rejection_note": rejectionNote.String,
		})
	}
	return result, nil
}

// ADMIN: Ambil semua achievement dengan filtering & pagination
func (r *AchievementRepository) AdminGetAll(
	ctx context.Context,
	f AchievementAdminFilter,
) ([]map[string]interface{}, int64, error) {
	base := `
        SELECT 
            id,
            student_uuid,
            mongo_achievement_id,
            status,
            submitted_at,
            verified_at,
            verified_by,
            rejection_note,
            created_at,
            updated_at
        FROM achievement_references
        WHERE is_deleted = FALSE
    `
	args := []interface{}{}
	idx := 1

	// Filter status
	if f.Status != "" {
		base += fmt.Sprintf(" AND status = $%d", idx)
		args = append(args, f.Status)
		idx++
	}

	// Filter student_uuid
	if f.StudentID != "" {
		base += fmt.Sprintf(" AND student_uuid = $%d", idx)
		args = append(args, f.StudentID)
		idx++
	}

	// ----- Hitung total (untuk pagination) -----
	countQuery := "SELECT COUNT(*) FROM (" + base + ") AS sub"
	var total int64
	if err := r.DB.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// ----- Sorting -----
	sortCol := "created_at"
	switch f.Sort {
	case "created_at", "submitted_at", "verified_at", "status":
		sortCol = f.Sort
	}
	orderDir := "DESC"
	if strings.ToLower(f.Order) == "asc" {
		orderDir = "ASC"
	}
	base += " ORDER BY " + sortCol + " " + orderDir

	// ----- Pagination (LIMIT / OFFSET) -----
	if f.Limit > 0 {
		base += fmt.Sprintf(" LIMIT $%d", idx)
		args = append(args, f.Limit)
		idx++
	}
	if f.Offset > 0 {
		base += fmt.Sprintf(" OFFSET $%d", idx)
		args = append(args, f.Offset)
		idx++
	}

	// ----- Eksekusi query utama -----
	rows, err := r.DB.QueryContext(ctx, base, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var results []map[string]interface{}
	for rows.Next() {
		var (
			id            string
			studentID     sql.NullString
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
			&id,
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
			return nil, 0, err
		}
		row := map[string]interface{}{
			"id":             id,
			"student_uuid":   nil,
			"mongo_id":       mongoID,
			"status":         status,
			"submitted_at":   nil,
			"verified_at":    nil,
			"verified_by":    nil,
			"rejection_note": nil,
			"created_at":     createdAt,
			"updated_at":     updatedAt,
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
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return results, total, nil
}

// Ambil semua student_uuid yang dibimbing dosen tertentu
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
            student_uuid,
            mongo_achievement_id,
            status,
            submitted_at,
            verified_at,
            verified_by,
            rejection_note,
            created_at,
            updated_at
        FROM achievement_references
        WHERE student_uuid = ANY($1)
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
			"student_uuid":   studentID,
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

// Verifikasi achievement dan update poin mahasiswa
func (r *AchievementRepository) Verify(
	ctx context.Context,
	achievementID string,
	studentID string,
	lecturerID int64,
	points float64,
) error {

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// 1. Update status → verified
	_, err = tx.ExecContext(ctx, `
        UPDATE achievement_references
        SET status = 'verified',
            verified_at = NOW(),
            verified_by = $1,
            updated_at = NOW()
        WHERE mongo_achievement_id = $2
        AND student_uuid = $3
        AND status = 'submitted'
    `, lecturerID, achievementID, studentID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 2. Tambah poin mahasiswa
	_, err = tx.ExecContext(ctx, `
        UPDATE students
        SET points = points + $1
        WHERE id = $2
    `, points, studentID)
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (r *AchievementRepository) GetStudentIDByAchievement(ctx context.Context, achievementID string) (string, error) {
	query := `
        SELECT student_uuid
        FROM achievement_references
        WHERE mongo_achievement_id = $1
        LIMIT 1
    `
	var studentID string
	err := r.DB.QueryRowContext(ctx, query, achievementID).Scan(&studentID)
	return studentID, err
}

func (r *AchievementRepository) IsStudentSupervised(ctx context.Context, lecturerID int64, studentID string) (bool, error) {
	query := `
        SELECT COUNT(*)
        FROM students
        WHERE id = $1 AND advisor_id = $2
    `
	var count int
	err := r.DB.QueryRowContext(ctx, query, studentID, lecturerID).Scan(&count)
	return count > 0, err
}

// Reject achievement
func (r *AchievementRepository) Reject(
	ctx context.Context,
	achievementID string,
	studentID string,
	lecturerID int64,
	note string,
) error {
	query := `
        UPDATE achievement_references
        SET status = 'rejected',
            rejection_note = $1,
            verified_by = $2, 
            updated_at = NOW()
        WHERE mongo_achievement_id = $3
          AND student_uuid = $4
          AND status = 'submitted'
    `
	_, err := r.DB.ExecContext(ctx, query, note, lecturerID, achievementID, studentID)
	return err
}

// Soft delete: update status menjadi deleted
func (r *AchievementRepository) SoftDelete(ctx context.Context, achievementID string, userID int64) error {
	query := `
        UPDATE achievement_references
        SET is_deleted = true, updated_at = NOW()
        WHERE mongo_achievement_id = $1
        AND student_id = (
            SELECT id FROM students WHERE user_id = $2
        )
        AND status = 'draft';
    `
	_, err := r.DB.ExecContext(ctx, query, achievementID, userID)
	return err
}

// GET reference by mongo_achievement_id
func (r *AchievementRepository) GetReferenceByMongoID(
	ctx context.Context,
	mongoID string,
) (map[string]interface{}, error) {
	query := `
		SELECT 
			id,
			student_uuid,
			mongo_achievement_id,
			status,
			submitted_at,
			verified_at,
			verified_by,
			rejection_note,
			created_at,
			updated_at
		FROM achievement_references
		WHERE mongo_achievement_id = $1
		  AND is_deleted = FALSE
		LIMIT 1
	`
	row := r.DB.QueryRowContext(ctx, query, mongoID)
	var (
		id            string
		studentUUID   sql.NullString
		status        string
		submittedAt   sql.NullTime
		verifiedAt    sql.NullTime
		verifiedBy    sql.NullInt64
		rejectionNote sql.NullString
		createdAt     time.Time
		updatedAt     time.Time
	)
	if err := row.Scan(
		&id,
		&studentUUID,
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
	ref := map[string]interface{}{
		"id":             id,
		"mongo_id":       mongoID,
		"status":         status,
		"student_uuid":   nil,
		"submitted_at":   nil,
		"verified_at":    nil,
		"verified_by":    nil,
		"rejection_note": nil,
		"created_at":     createdAt,
		"updated_at":     updatedAt,
	}
	if studentUUID.Valid {
		ref["student_uuid"] = studentUUID.String
	}
	if submittedAt.Valid {
		ref["submitted_at"] = submittedAt.Time
	}
	if verifiedAt.Valid {
		ref["verified_at"] = verifiedAt.Time
	}
	if verifiedBy.Valid {
		ref["verified_by"] = verifiedBy.Int64
	}
	if rejectionNote.Valid {
		ref["rejection_note"] = rejectionNote.String
	}
	return ref, nil
}

// GET achievement history by mongo_achievement_id
func (r *AchievementRepository) GetHistoryByMongoID(
	ctx context.Context,
	mongoID string,
) ([]map[string]interface{}, error) {
	query := `
		SELECT 
			status,
			created_at,
			submitted_at,
			verified_at,
			verified_by,
			rejection_note
		FROM achievement_references
		WHERE mongo_achievement_id = $1
		  AND is_deleted = FALSE
		LIMIT 1
	`
	var (
		status        string
		createdAt     time.Time
		submittedAt   sql.NullTime
		verifiedAt    sql.NullTime
		verifiedBy    sql.NullInt64
		rejectionNote sql.NullString
	)
	err := r.DB.QueryRowContext(ctx, query, mongoID).Scan(
		&status,
		&createdAt,
		&submittedAt,
		&verifiedAt,
		&verifiedBy,
		&rejectionNote,
	)
	if err != nil {
		return nil, err
	}
	var history []map[string]interface{}

	// Draft
	history = append(history, map[string]interface{}{
		"status": "draft",
		"at":     createdAt,
		"by":     nil,
	})

	// Submitted
	if submittedAt.Valid {
		history = append(history, map[string]interface{}{
			"status": "submitted",
			"at":     submittedAt.Time,
			"by":     "student",
		})
	}

	// Verified / Rejected
	if status == "verified" && verifiedAt.Valid {
		history = append(history, map[string]interface{}{
			"status": "verified",
			"at":     verifiedAt.Time,
			"by":     verifiedBy.Int64,
		})
	}
	if status == "rejected" && verifiedAt.Valid {
		history = append(history, map[string]interface{}{
			"status": "rejected",
			"at":     verifiedAt.Time,
			"by":     verifiedBy.Int64,
			"note":   rejectionNote.String,
		})
	}
	return history, nil
}

// GET all verified achievement references
func (r *AchievementRepository) GetVerifiedAchievementRefs(
	ctx context.Context,
) ([]map[string]interface{}, error) {
	query := `
		SELECT
			ar.mongo_achievement_id,
			ar.verified_at
		FROM achievement_references ar
		WHERE ar.status = 'verified'
		  AND ar.is_deleted = false
	`
	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []map[string]interface{}
	for rows.Next() {
		var mongoID string
		var verifiedAt time.Time
		if err := rows.Scan(&mongoID, &verifiedAt); err != nil {
			return nil, err
		}
		results = append(results, map[string]interface{}{
			"mongo_id":   mongoID,
			"verified_at": verifiedAt,
		})
	}
	return results, nil
}

// GET top 5 students dengan achievement VERIFIED terbanyak
func (r *AchievementRepository) GetTopStudents(
	ctx context.Context,
) ([]map[string]interface{}, error) {
	query := `
		SELECT student_uuid, COUNT(*) AS total
		FROM achievement_references
		WHERE status = 'verified'
			AND is_deleted = false
			AND student_uuid IS NOT NULL
		GROUP BY student_uuid
		ORDER BY total DESC
		LIMIT 5
	`
	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []map[string]interface{}
	for rows.Next() {
		var studentID string
		var total int
		if err := rows.Scan(&studentID, &total); err != nil {
			return nil, err
		}
		results = append(results, map[string]interface{}{
			"student_id": studentID,
			"total":      total,
		})
	}
	return results, nil
}

// GET achievements by student UUID
func (r *AchievementRepository) GetByStudentUUID(
	ctx context.Context,
	studentUUID string,
) ([]map[string]interface{}, error) {

	query := `
		SELECT
			mongo_achievement_id,
			status,
			verified_at
		FROM achievement_references
		WHERE student_uuid = $1
		  AND is_deleted = false
	`
	rows, err := r.DB.QueryContext(ctx, query, studentUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []map[string]interface{}
	for rows.Next() {
		var mongoID string
		var status string
		var verifiedAt *string
		if err := rows.Scan(&mongoID, &status, &verifiedAt); err != nil {
			return nil, err
		}
		results = append(results, map[string]interface{}{
			"mongo_id":    mongoID,
			"status":      status,
			"verified_at": verifiedAt,
		})
	}
	return results, nil
}
