package mocks

import (
	"context"

	"uas/app/model"

	"github.com/stretchr/testify/mock"
)

type MockStudentRepository struct {
	mock.Mock
}

// =======================
// METHODS DIPAKAI STUDENT SERVICE
// =======================

func (m *MockStudentRepository) GetAll(
	ctx context.Context,
) ([]map[string]interface{}, error) {
	args := m.Called(ctx)
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func (m *MockStudentRepository) GetByID(
	ctx context.Context,
	studentID string,
) (map[string]interface{}, error) {
	args := m.Called(ctx, studentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockStudentRepository) LecturerExists(
	ctx context.Context,
	lecturerID int64,
) (bool, error) {
	args := m.Called(ctx, lecturerID)
	return args.Bool(0), args.Error(1)
}

func (m *MockStudentRepository) AssignAdvisor(
	ctx context.Context,
	studentID string,
	lecturerID int64,
) error {
	args := m.Called(ctx, studentID, lecturerID)
	return args.Error(0)
}

func (m *MockStudentRepository) GetStudentAchievements(
	ctx context.Context,
	studentID string,
) ([]map[string]interface{}, error) {
	args := m.Called(ctx, studentID)
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

// =======================
// METHODS DIPAKAI ADMIN SERVICE
// =======================

func (m *MockStudentRepository) Create(
	ctx context.Context,
	s model.StudentCreate,
) (string, error) {
	args := m.Called(ctx, s)
	return args.String(0), args.Error(1)
}

func (m *MockStudentRepository) UpdateProfile(
	ctx context.Context,
	s model.StudentCreate,
) error {
	args := m.Called(ctx, s)
	return args.Error(0)
}

func (m *MockStudentRepository) DeleteByUserID(
	ctx context.Context,
	userID int64,
) (bool, error) {
	args := m.Called(ctx, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockStudentRepository) GetStudentID(
	ctx context.Context,
	userID int64,
) (string, error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.Error(1)
}

// =======================
// DUMMY (NOT USED IN TEST)
// =======================

func (m *MockStudentRepository) GetStudentsByAdvisor(
	ctx context.Context,
	advisorID int64,
) ([]string, error) {
	return []string{}, nil
}
