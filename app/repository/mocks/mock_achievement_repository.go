package mocks

import (
	"context"

	"uas/app/repository"

	"github.com/stretchr/testify/mock"
)

// MockAchievementRepository
// Digunakan untuk unit test service (tanpa DB)
type MockAchievementRepository struct {
	mock.Mock
}

// =======================
// METHODS YANG DIPAKAI SERVICE
// =======================

func (m *MockAchievementRepository) GetStudentID(
	ctx context.Context,
	userID int64,
) (string, error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.Error(1)
}

func (m *MockAchievementRepository) InsertReference(
	ctx context.Context,
	studentID string,
	mongoID string,
) error {
	args := m.Called(ctx, studentID, mongoID)
	return args.Error(0)
}

func (m *MockAchievementRepository) Submit(
	ctx context.Context,
	achievementID string,
	studentID string,
) error {
	args := m.Called(ctx, achievementID, studentID)
	return args.Error(0)
}

func (m *MockAchievementRepository) GetStudentIDByAchievement(
	ctx context.Context,
	achievementID string,
) (string, error) {
	args := m.Called(ctx, achievementID)
	return args.String(0), args.Error(1)
}

func (m *MockAchievementRepository) GetLecturerID(
	ctx context.Context,
	userID int64,
) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockAchievementRepository) IsStudentSupervised(
	ctx context.Context,
	lecturerID int64,
	studentID string,
) (bool, error) {
	args := m.Called(ctx, lecturerID, studentID)
	return args.Bool(0), args.Error(1)
}

func (m *MockAchievementRepository) Verify(
	ctx context.Context,
	achievementID string,
	studentID string,
	lecturerID int64,
	points float64,
) error {
	args := m.Called(ctx, achievementID, studentID, lecturerID, points)
	return args.Error(0)
}

func (m *MockAchievementRepository) Reject(
	ctx context.Context,
	achievementID string,
	studentID string,
	lecturerID int64,
	note string,
) error {
	args := m.Called(ctx, achievementID, studentID, lecturerID, note)
	return args.Error(0)
}

func (m *MockAchievementRepository) SoftDelete(
	ctx context.Context,
	achievementID string,
	userID int64,
) error {
	args := m.Called(ctx, achievementID, userID)
	return args.Error(0)
}

// =======================
// DUMMY METHODS (WAJIB ADA)
// =======================

func (m *MockAchievementRepository) GetByStudentID(
	ctx context.Context,
	studentID string,
	includeDeleted bool,
) ([]map[string]interface{}, error) {
	return []map[string]interface{}{}, nil
}

func (m *MockAchievementRepository) AdminGetAll(
	ctx context.Context,
	f repository.AchievementAdminFilter,
) ([]map[string]interface{}, int64, error) {
	return []map[string]interface{}{}, 0, nil
}

func (m *MockAchievementRepository) GetStudentsByAdvisor(
	ctx context.Context,
	advisorID int64,
) ([]string, error) {
	return []string{}, nil
}

func (m *MockAchievementRepository) GetReferencesByStudentList(
	ctx context.Context,
	studentIDs []string,
) ([]map[string]interface{}, error) {
	return []map[string]interface{}{}, nil
}

func (m *MockAchievementRepository) GetReferenceByMongoID(
	ctx context.Context,
	mongoID string,
) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}

func (m *MockAchievementRepository) GetHistoryByMongoID(
	ctx context.Context,
	mongoID string,
) ([]map[string]interface{}, error) {
	return []map[string]interface{}{}, nil
}

func (m *MockAchievementRepository) GetVerifiedAchievementRefs(
	ctx context.Context,
) ([]map[string]interface{}, error) {
	return []map[string]interface{}{}, nil
}

func (m *MockAchievementRepository) GetTopStudents(
	ctx context.Context,
) ([]map[string]interface{}, error) {
	return []map[string]interface{}{}, nil
}

func (m *MockAchievementRepository) GetByStudentUUID(
	ctx context.Context,
	studentUUID string,
) ([]map[string]interface{}, error) {
	return []map[string]interface{}{}, nil
}
