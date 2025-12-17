package service

import (
	"database/sql"

	"uas/app/repository"
	"uas/utils"

	"github.com/gofiber/fiber/v2"
)

type StudentService struct {
	Repo repository.StudentRepository
}

type AssignAdvisorRequest struct {
	AdvisorID int64 `json:"advisor_id"`
}

func NewStudentService(repo repository.StudentRepository) *StudentService {
	return &StudentService{Repo: repo}
}

// ADMIN: GET /students
func (s *StudentService) GetAll(c *fiber.Ctx) error {
	// AdminOnly sudah di route
	_ = c.Locals("claims").(*utils.Claims)
	students, err := s.Repo.GetAll(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed_get_students",
		})
	}
	return c.JSON(fiber.Map{
		"data": students,
	})
}

// ADMIN: GET /students/:id
func (s *StudentService) GetByID(c *fiber.Ctx) error {
	_ = c.Locals("claims").(*utils.Claims)
	studentID := c.Params("id")
	if studentID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid_student_id",
		})
	}
	student, err := s.Repo.GetByID(c.Context(), studentID)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{
				"error": "student_not_found",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error": "failed_get_student",
		})
	}
	return c.JSON(student)
}

// ADMIN: PUT /students/:id/advisor
func (s *StudentService) AssignAdvisor(c *fiber.Ctx) error {
	_ = c.Locals("claims").(*utils.Claims)
	studentID := c.Params("id")
	if studentID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid_student_id",
		})
	}
	var req AssignAdvisorRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid_request",
		})
	}
	if req.AdvisorID <= 0 {
		return c.Status(422).JSON(fiber.Map{
			"error": "invalid_advisor_id",
		})
	}

	// 1. Pastikan student ada
	_, err := s.Repo.GetByID(c.Context(), studentID)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{
				"error": "student_not_found",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error": "failed_get_student",
		})
	}

	// 2. Pastikan lecturer ada
	exists, err := s.Repo.LecturerExists(c.Context(), req.AdvisorID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed_check_lecturer",
		})
	}
	if !exists {
		return c.Status(404).JSON(fiber.Map{
			"error": "lecturer_not_found",
		})
	}

	// 3. Assign advisor
	if err := s.Repo.AssignAdvisor(c.Context(), studentID, req.AdvisorID); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed_assign_advisor",
		})
	}
	return c.JSON(fiber.Map{
		"message":     "advisor_assigned",
		"student_id":  studentID,
		"advisor_id":  req.AdvisorID,
	})
}

// ADMIN: GET /students/:id/achievements
func (s *StudentService) GetAchievements(c *fiber.Ctx) error {
	_ = c.Locals("claims").(*utils.Claims)
	ctx := c.Context()
	studentID := c.Params("id")
	if studentID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid_student_id",
		})
	}

	// 1. Pastikan student ada
	_, err := s.Repo.GetByID(ctx, studentID)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{
				"error": "student_not_found",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error": "failed_get_student",
		})
	}

	// 2. Ambil achievement references (Postgres)
	refs, err := s.Repo.GetStudentAchievements(ctx, studentID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed_get_achievements",
		})
	}
	return c.JSON(fiber.Map{
		"student_id":  studentID,
		"achievements": refs,
	})
}
