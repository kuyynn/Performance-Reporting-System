package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"uas/app/repository/mocks"
	"uas/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupLecturerApp(service *LecturerService) *fiber.App {
	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("claims", &utils.Claims{
			UserID: 1,
			Role:   "dosen wali",
		})
		return c.Next()
	})

	app.Get("/lecturers", service.GetAll)
	app.Get("/lecturers/:id/advisees", service.GetAdvisees)
	return app
}

func TestLecturer_GetAll_Success(t *testing.T) {
	repo := new(mocks.MockLecturerRepository)
	service := NewLecturerService(repo)
	app := setupLecturerApp(service)

	repo.On("GetAll", mock.Anything).
		Return([]map[string]interface{}{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/lecturers", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	repo.AssertExpectations(t)
}

func TestLecturer_GetAdvisees_Success(t *testing.T) {
	repo := new(mocks.MockLecturerRepository)
	service := NewLecturerService(repo)
	app := setupLecturerApp(service)

	repo.On("GetLecturerIDByUserID", mock.Anything, int64(1)).
		Return(int64(5), nil)

	repo.On("GetAdvisees", mock.Anything, int64(5)).
		Return([]map[string]interface{}{}, nil)

	req := httptest.NewRequest(
		http.MethodGet,
		"/lecturers/5/advisees",
		nil,
	)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	repo.AssertExpectations(t)
}
