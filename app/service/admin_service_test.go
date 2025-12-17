package service

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"uas/app/model"
	"uas/app/repository/mocks"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// =======================
// HELPER
// =======================

func setupApp(service *AdminService) *fiber.App {
	app := fiber.New()
	app.Post("/users", service.CreateUser)
	app.Delete("/users/:id", service.DeleteUser)
	return app
}

// =======================
// CREATE USER
// =======================

func TestCreateUser_InvalidRole(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	studentRepo := new(mocks.MockStudentRepository)
	lecturerRepo := new(mocks.MockLecturerRepository)

	service := NewAdminService(userRepo, studentRepo, lecturerRepo)
	app := setupApp(service)

	input := model.AdminCreateUserRequest{
		Username: "test",
		Role:     "superadmin",
	}

	body, _ := json.Marshal(input)
	req := httptest.NewRequest(
		http.MethodPost,
		"/users",
		bytes.NewBuffer(body),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestCreateUser_AdminSuccess(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	studentRepo := new(mocks.MockStudentRepository)
	lecturerRepo := new(mocks.MockLecturerRepository)

	service := NewAdminService(userRepo, studentRepo, lecturerRepo)
	app := setupApp(service)

	input := model.AdminCreateUserRequest{
		Username: "admin1",
		FullName: "Admin One",
		Email:    "admin@test.com",
		Password: "secret",
		Role:     "admin",
	}

	userRepo.On("GetRoleIDByName", "admin").
		Return("role-admin", nil)

	userRepo.On(
		"CreateRaw",
		mock.Anything,
		"admin1",
		"Admin One",
		"admin@test.com",
		mock.Anything,
		"role-admin",
	).Return(int64(10), nil)

	body, _ := json.Marshal(input)
	req := httptest.NewRequest(
		http.MethodPost,
		"/users",
		bytes.NewBuffer(body),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	userRepo.AssertExpectations(t)
}

// =======================
// DELETE USER
// =======================

func TestDeleteUser_Success(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	studentRepo := new(mocks.MockStudentRepository)
	lecturerRepo := new(mocks.MockLecturerRepository)

	service := NewAdminService(userRepo, studentRepo, lecturerRepo)
	app := setupApp(service)

	userRepo.On("Delete", mock.Anything, int64(1)).
		Return(nil)

	studentRepo.On("GetStudentID", mock.Anything, int64(1)).
		Return("", fiber.ErrNotFound)

	lecturerRepo.On("GetLecturerID", mock.Anything, int64(1)).
		Return(int64(0), fiber.ErrNotFound)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/users/1",
		nil,
	)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	userRepo.AssertExpectations(t)
}
