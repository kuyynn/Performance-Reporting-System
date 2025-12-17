package service

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"uas/app/repository/mocks"
	"uas/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupStudentApp(service *StudentService) *fiber.App {
	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("claims", &utils.Claims{
			UserID: 1,
			Role:   "admin",
		})
		return c.Next()
	})

	app.Get("/students", service.GetAll)
	app.Get("/students/:id", service.GetByID)
	app.Put("/students/:id/advisor", service.AssignAdvisor)
	app.Get("/students/:id/achievements", service.GetAchievements)
	return app
}

func TestStudent_GetAll_Success(t *testing.T) {
	repo := new(mocks.MockStudentRepository)
	service := NewStudentService(repo)
	app := setupStudentApp(service)

	repo.On("GetAll", mock.Anything).
		Return([]map[string]interface{}{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/students", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	repo.AssertExpectations(t)
}

func TestStudent_GetByID_NotFound(t *testing.T) {
	repo := new(mocks.MockStudentRepository)
	service := NewStudentService(repo)
	app := setupStudentApp(service)

	repo.On("GetByID", mock.Anything, "abc").
		Return(nil, sql.ErrNoRows)

	req := httptest.NewRequest(http.MethodGet, "/students/abc", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)

	repo.AssertExpectations(t)
}

func TestStudent_AssignAdvisor_Success(t *testing.T) {
	repo := new(mocks.MockStudentRepository)
	service := NewStudentService(repo)
	app := setupStudentApp(service)

	body, _ := json.Marshal(map[string]int64{
		"advisor_id": 10,
	})

	repo.On("GetByID", mock.Anything, "stu-1").
		Return(map[string]interface{}{}, nil)

	repo.On("LecturerExists", mock.Anything, int64(10)).
		Return(true, nil)

	repo.On("AssignAdvisor", mock.Anything, "stu-1", int64(10)).
		Return(nil)

	req := httptest.NewRequest(
		http.MethodPut,
		"/students/stu-1/advisor",
		bytes.NewBuffer(body),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	repo.AssertExpectations(t)
}

func TestStudent_GetAchievements_Success(t *testing.T) {
	repo := new(mocks.MockStudentRepository)
	service := NewStudentService(repo)
	app := setupStudentApp(service)

	repo.On("GetByID", mock.Anything, "stu-1").
		Return(map[string]interface{}{}, nil)

	repo.On("GetStudentAchievements", mock.Anything, "stu-1").
		Return([]map[string]interface{}{}, nil)

	req := httptest.NewRequest(
		http.MethodGet,
		"/students/stu-1/achievements",
		nil,
	)

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	repo.AssertExpectations(t)
}
