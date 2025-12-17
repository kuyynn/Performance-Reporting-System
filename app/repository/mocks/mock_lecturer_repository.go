package mocks

import (
	"context"

	"uas/app/model"

	"github.com/stretchr/testify/mock"
)

type MockLecturerRepository struct {
	mock.Mock
}

// =======================
// METHODS DIPAKAI LECTURER SERVICE
// =======================

func (m *MockLecturerRepository) GetAll(
	ctx context.Context,
) ([]map[string]interface{}, error) {
	args := m.Called(ctx)
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func (m *MockLecturerRepository) GetAdvisees(
	ctx context.Context,
	lecturerID int64,
) ([]map[string]interface{}, error) {
	args := m.Called(ctx, lecturerID)
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func (m *MockLecturerRepository) GetLecturerIDByUserID(
	ctx context.Context,
	userID int64,
) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

// =======================
// METHODS DIPAKAI ADMIN SERVICE
// =======================

func (m *MockLecturerRepository) Create(
	ctx context.Context,
	l model.LecturerCreate,
) (int64, error) {
	args := m.Called(ctx, l)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockLecturerRepository) Update(
	ctx context.Context,
	l model.LecturerCreate,
) error {
	args := m.Called(ctx, l)
	return args.Error(0)
}

func (m *MockLecturerRepository) DeleteByUserID(
	ctx context.Context,
	userID int64,
) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockLecturerRepository) GetLecturerID(
	ctx context.Context,
	userID int64,
) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}
