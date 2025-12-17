package service

import (
	"strconv"

	"uas/app/repository"
	"uas/utils"

	"github.com/gofiber/fiber/v2"
)

type LecturerService struct {
	Repo repository.LecturerRepository
}

func NewLecturerService(repo repository.LecturerRepository) *LecturerService {
	return &LecturerService{Repo: repo}
}

// ADMIN: GET /lecturers
func (s *LecturerService) GetAll(c *fiber.Ctx) error {
	_ = c.Locals("claims").(*utils.Claims)
	lecturers, err := s.Repo.GetAll(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed_get_lecturers",
		})
	}
	return c.JSON(fiber.Map{
		"data": lecturers,
	})
}

// DOSEN WALI: GET /lecturers/:id/advisees
func (s *LecturerService) GetAdvisees(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*utils.Claims)
	if claims.Role != "dosen wali" {
		return c.Status(403).JSON(fiber.Map{
			"error": "lecturer_only",
		})
	}
	lecturerID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid_lecturer_id",
		})
	}

	// cek apakah lecturerID milik dosen yang sedang login
	myLecturerID, err := s.Repo.GetLecturerIDByUserID(c.Context(), claims.UserID)
	if err != nil || myLecturerID != lecturerID {
		return c.Status(403).JSON(fiber.Map{
			"error": "forbidden",
		})
	}
	students, err := s.Repo.GetAdvisees(c.Context(), lecturerID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed_get_advisees",
		})
	}
	return c.JSON(fiber.Map{
		"lecturer_id": lecturerID,
		"students":    students,
	})
}
