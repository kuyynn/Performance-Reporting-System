package service

import (
	"uas/app/model"
	"uas/app/repository"
	"uas/utils"

	"github.com/gofiber/fiber/v2"
)

type AdminService struct {
	UserRepo     repository.UserRepository
	StudentRepo  repository.StudentRepository
	LecturerRepo repository.LecturerRepository
}

// CONSTRUCTOR
func NewAdminService(userRepo repository.UserRepository,
	studentRepo repository.StudentRepository,
	lecturerRepo repository.LecturerRepository) *AdminService {
	return &AdminService{
		UserRepo:     userRepo,
		StudentRepo:  studentRepo,
		LecturerRepo: lecturerRepo,
	}
}

// CREATE USER
func (s *AdminService) CreateUser(c *fiber.Ctx) error {
	var input model.AdminCreateUserRequest
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid_request"})
	}

	// --- VALIDASI ROLE ---
	if input.Role != "mahasiswa" && input.Role != "dosen wali" && input.Role != "admin" {
		return c.Status(400).JSON(fiber.Map{"error": "invalid_role"})
	}

	// --- HASH PASSWORD ---
	passHash, err := utils.HashPassword(input.Password)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "password_hash_failed"})
	}

	// --- AMBIL ROLE ID dari table roles ---
	roleID, err := s.UserRepo.GetRoleIDByName(input.Role)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "role_not_found"})
	}

	// --- INSERT USER ---
	userID, err := s.UserRepo.CreateRaw(
		c.Context(),
		input.Username,
		input.FullName,
		input.Email,
		passHash,
		roleID,
	)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// ROLE = MAHASISWA → buat profile
	if input.Role == "mahasiswa" {

		// generate NIM: 43420 + user_id
		nim := "4342025" + utils.IntToString(userID)

		student := model.StudentCreate{
			UserID:       userID,
			ProgramStudy: input.ProgramStudy,
			AcademicYear: input.AcademicYear,
			AdvisorID:    input.AdvisorID, // boleh null
			Nim:          nim,
		}

		_, err := s.StudentRepo.Create(c.Context(), student)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed_create_student_profile"})
		}

		return c.JSON(fiber.Map{
			"message": "student user created",
			"user_id": userID,
			"nim":     nim,
		})
	}

	// ROLE = DOSEN WALI → buat profile
	if input.Role == "dosen wali" {

		if input.NIP == "" || input.Department == "" {
			return c.Status(400).JSON(fiber.Map{"error": "nip_and_department_required"})
		}

		lecturer := model.LecturerCreate{
			UserID:     userID,
			NIP:        input.NIP,
			Department: input.Department,
		}

		_, err := s.LecturerRepo.Create(c.Context(), lecturer)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed_create_lecturer_profile"})
		}

		return c.JSON(fiber.Map{
			"message": "lecturer user created",
			"user_id": userID,
			"nip":     input.NIP,
		})
	}

	// ROLE = ADMIN → hanya insert user
	return c.JSON(fiber.Map{
		"message": "admin user created",
		"user_id": userID,
	})
}

// UPDATE USER
func (s *AdminService) UpdateUser(c *fiber.Ctx) error {
	userID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid_user_id"})
	}

	var input model.AdminUpdateUserRequest
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid_request"})
	}

	// VALIDASI ROLE
	if input.Role != "admin" && input.Role != "mahasiswa" && input.Role != "dosen wali" {
		return c.Status(400).JSON(fiber.Map{"error": "invalid_role"})
	}

	// GET ROLE ID
	roleID, err := s.UserRepo.GetRoleIDByName(input.Role)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "role_not_found"})
	}

	// UPDATE USERS
	err = s.UserRepo.UpdateRaw(c.Context(), int64(userID),
		input.Username,
		input.FullName,
		input.Email,
		roleID,
	)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// UPDATE MAHASISWA PROFILE
	if input.Role == "mahasiswa" {
		err = s.StudentRepo.UpdateProfile(
			c.Context(),
			model.StudentCreate{
				UserID:       int64(userID),
				ProgramStudy: input.ProgramStudy,
				AcademicYear: input.AcademicYear,
				AdvisorID:    input.AdvisorID,
			},
		)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed_update_student"})
		}
	}

	// UPDATE DOSEN PROFILE
	if input.Role == "dosen wali" {
		err = s.LecturerRepo.Update(c.Context(),
			model.LecturerCreate{
				UserID:     int64(userID),
				NIP:        input.NIP,
				Department: input.Department,
			},
		)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed_update_lecturer"})
		}
	}
	return c.JSON(fiber.Map{
		"message": "user updated",
		"user_id": userID,
	})
}

// DELETE USER
func (s *AdminService) DeleteUser(c *fiber.Ctx) error {
    userID, err := c.ParamsInt("id")
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "invalid_user_id"})
    }
    ctx := c.Context()
    // --- Cek apakah user memiliki student profile ---
    _, err = s.StudentRepo.GetStudentID(ctx, int64(userID))
    if err == nil {
        // Ada student profile → hapus mahasiswa
        _, _ = s.StudentRepo.DeleteByUserID(ctx, int64(userID))
    }
    // --- Cek apakah user memiliki lecturer profile ---
    _, err = s.LecturerRepo.GetLecturerID(ctx, int64(userID))
    if err == nil {
        // Ada lecturer profile → hapus dosen
        _ = s.LecturerRepo.DeleteByUserID(ctx, int64(userID))
    }
    // --- Hapus user dari table users ---
    err = s.UserRepo.Delete(ctx, int64(userID))
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "delete_failed"})
    }
    return c.JSON(fiber.Map{
        "message": "user deleted",
        "user_id": userID,
    })
}

// GET ALL USERS
func (s *AdminService) GetAllUsers(c *fiber.Ctx) error {
	users, err := s.UserRepo.FindAll(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed_fetch_users"})
	}
	return c.JSON(users)
}
