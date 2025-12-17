package mocks

import (
	"context"

	"uas/app/model"

	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

// ===== METHODS YANG DIPAKAI DI ADMIN SERVICE =====

func (m *MockUserRepository) GetRoleIDByName(roleName string) (string, error) {
	args := m.Called(roleName)
	return args.String(0), args.Error(1)
}

func (m *MockUserRepository) CreateRaw(
	ctx context.Context,
	username, fullName, email, passHash, roleID string,
) (int64, error) {
	args := m.Called(ctx, username, fullName, email, passHash, roleID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) Delete(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateRaw(
	ctx context.Context,
	id int64,
	username, fullName, email, roleID string,
) error {
	args := m.Called(ctx, id, username, fullName, email, roleID)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateRole(
	ctx context.Context,
	userID int64,
	roleID string,
) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}

// ===== DUMMY METHODS (TIDAK DIPAKAI, TAPI WAJIB ADA) =====
func (m *MockUserRepository) Create(
	ctx context.Context,
	user model.UserCreateRequest,
) (int64, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) Update(
	ctx context.Context,
	user model.UserUpdateRequest,
) error {
	return nil
}

func (m *MockUserRepository) FindAll(
	ctx context.Context,
) (*[]model.UserResponse, error) {
	return &[]model.UserResponse{}, nil
}

func (m *MockUserRepository) FindByUsernameOrEmail(
	usernameOrEmail string,
) (*model.User, error) {
	args := m.Called(usernameOrEmail)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) GetPermissionsByUserID(
	userID int64,
) ([]string, error) {
	args := m.Called(userID)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockUserRepository) FindById(
	ctx context.Context,
	userID int64,
) (*model.UserResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.UserResponse), args.Error(1)
}

func (m *MockUserRepository) Logout() {
}
