package routes

import (
	"uas/app/service"
	"uas/middleware"

	"github.com/gofiber/fiber/v2"
	swagger "github.com/gofiber/swagger"
)

func SetupRoutes(
	app *fiber.App,
	userService service.UserService,
	authService *service.AuthService,
	achievementService *service.AchievementService,
	adminService *service.AdminService,
	authMiddleware fiber.Handler,
	studentService *service.StudentService,
	lecturerService *service.LecturerService,
	reportService *service.ReportService,
) {

	// PUBLIC AUTH (NO TOKEN)
	auth := app.Group("/api/v1/auth")

	auth.Post("/login", authService.LoginHandler)
	auth.Post("/refresh", authService.RefreshHandler)

	// PROTECTED ROUTES (JWT)
	api := app.Group("/api/v1", authMiddleware)

	// AUTH PROTECTED
	api.Get("/auth/profile", authService.ProfileHandler)
	api.Post("/auth/logout", authService.LogoutHandler)

	// ADMIN PROTECTED
	admin := api.Group("/admin", middleware.AdminOnly)

	// ADMIN: USER CRUD
	admin.Post("/users", adminService.CreateUser)
	admin.Put("/users/:id", adminService.UpdateUser)
	admin.Delete("/users/:id", adminService.DeleteUser)
	admin.Get("/users", adminService.GetAllUsers)
	admin.Get("/users/:id", adminService.GetUserByID)
	admin.Put("/users/:id/role", adminService.UpdateUserRole)

	// ADMIN: STUDENT
	admin.Get("/students", studentService.GetAll)
	admin.Get("/students/:id", studentService.GetByID)
	admin.Put("/students/:id/advisor", studentService.AssignAdvisor)
	admin.Get("/students/:id/achievements", studentService.GetAchievements)

	// ADMIN — VIEW ALL ACHIEVEMENTS
	admin.Get("/achievements", achievementService.AdminListAchievements)

	// ADMIN: LECTURER
	admin.Get("/lecturers", lecturerService.GetAll)
	api.Get("/lecturers/:id/advisees", lecturerService.GetAdvisees)

	// ADMIN: REPORTS statistics
	admin.Get("/reports/statistics", reportService.GetStatistics)

	// ACHIEVEMENTS
	api.Post("/achievements", achievementService.CreateHandler)
	api.Post("/achievements/:id/submit", achievementService.Submit)
	api.Delete("/achievements/:id", achievementService.DeleteHandler)
	api.Get("/achievements/me", achievementService.GetMyAchievements)
	api.Get("/achievements/supervised", achievementService.GetSupervisedAchievements)
	api.Post("/achievements/:id/verify", achievementService.Verify)
	api.Post("/achievements/:id/reject", achievementService.Reject)
	api.Get("/achievements/:id", achievementService.GetDetail)
	api.Put("/achievements/:id", achievementService.UpdateDraft)
	api.Get("/achievements/:id/history", achievementService.GetHistory)
	api.Post("/achievements/:id/attachments", achievementService.UploadAttachment)
	api.Get("/reports/student/:id", reportService.GetStudentReport)

	// SWAGGER
	app.Get("/swagger/*", swagger.New(swagger.Config{
	URL: "/docs/swagger.yaml",
	}))
}
