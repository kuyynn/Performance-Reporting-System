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

func setupUserApp(service *UserService) *fiber.App {
	app := fiber.New()
	app.Post("/users", service.Create)
	return app
}

func TestUser_Create_Success(t *testing.T) {
	repo := new(mocks.MockUserRepository)
	service := NewUserService(repo)
	app := setupUserApp(&service)

	input := model.UserCreateRequest{
		Username: "test",
		Email:    "test@mail.com",
		Password: "secret",
		FullName: "Full Name",
	}

	repo.On(
		"Create",
		mock.Anything,
		mock.Anything,
	).Return(int64(1), nil)

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

	repo.AssertExpectations(t)
}

func TestUser_Create_InvalidBody(t *testing.T) {
	repo := new(mocks.MockUserRepository)
	service := NewUserService(repo)
	app := setupUserApp(&service)

	req := httptest.NewRequest(
		http.MethodPost,
		"/users",
		bytes.NewBuffer([]byte("invalid-json")),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}
