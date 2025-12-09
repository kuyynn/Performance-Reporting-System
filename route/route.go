package routes

import (
	"uas/app/service"

	"github.com/gofiber/fiber/v2"
)

// SetupRoutes mendaftarkan semua route API
func SetupRoutes(
	app *fiber.App,
	userService service.UserService,
	authService *service.AuthService,
	achievementService *service.AchievementService,
	authMiddleware fiber.Handler,
) {

	// ====== ROUTE PUBLIC ======

	// Login
	app.Post("/login", func(c *fiber.Ctx) error {
		var input service.LoginInput

		if err := c.BodyParser(&input); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "invalid_request",
			})
		}

		output, err := authService.Login(input)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(output)
	})

	// ROUTE PROTECTED (HARUS PAKAI JWT)
	api := app.Group("/api", authMiddleware)

	// USER CRUD
	api.Post("/users", userService.Create)
	api.Get("/users", userService.FindAll)
	api.Get("/users/:id", userService.FindById)
	api.Put("/users/:id", userService.Update)
	api.Delete("/users/:id", userService.Delete)

	// Logout
	api.Post("/logout", userService.Logout)

	// ACHIEVEMENT
	api.Post("/v1/achievements", func(c *fiber.Ctx) error {
		userID := c.Locals("userID").(int64)
		role := c.Locals("role").(string)
		var input service.AchievementInput
		if err := c.BodyParser(&input); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "invalid_request",
			})
		}
		ctx := c.Context()
		result, err := achievementService.CreateAchievement(ctx, userID, role, input)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.JSON(result)
	})

		// Mahasiswa submit prestasi (draft → submitted)
	api.Post("/v1/achievements/:id/submit", func(c *fiber.Ctx) error {
		userID := c.Locals("userID").(int64)
		role := c.Locals("role").(string)

		achievementID := c.Params("id")

		ctx := c.Context()
		err := achievementService.SubmitAchievement(ctx, userID, role, achievementID)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"message": "achievement submitted",
			"achievement_id": achievementID,
			"status": "submitted",
		})
	})
}
