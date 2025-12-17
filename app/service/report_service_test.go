package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"uas/app/repository"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func setupReportApp(service *ReportService) *fiber.App {
	app := fiber.New()
	app.Get("/reports/statistics", service.GetStatistics)
	return app
}

func TestReport_GetStatistics_EmptyData(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	achievementRepo := &repository.AchievementRepository{
		DB: db,
	}

	service := &ReportService{
		AchievementRepo:      achievementRepo,
		MongoAchievementRepo: &repository.MongoAchievementRepository{},
	}

	// 🔹 mock empty query result
	rows := sqlmock.NewRows([]string{"mongo_id"})
	mock.ExpectQuery("SELECT").
		WillReturnRows(rows)

	app := setupReportApp(service)

	req := httptest.NewRequest(
		http.MethodGet,
		"/reports/statistics",
		nil,
	)

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}
