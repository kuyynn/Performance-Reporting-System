package routes

import (
	"uas/app/service"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(
	app *fiber.App,
	userService service.UserService,
	authService *service.AuthService,
	achievementService *service.AchievementService,
	authMiddleware fiber.Handler,
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

	// USER CRUD
	api.Post("/users", userService.Create)
	api.Get("/users", userService.FindAll)
	api.Get("/users/:id", userService.FindById)
	api.Put("/users/:id", userService.Update)
	api.Delete("/users/:id", userService.Delete)

	// ACHIEVEMENTS
	api.Post("/v1/achievements", achievementService.Create)
	api.Post("/v1/achievements/:id/submit", achievementService.Submit)
	api.Delete("/v1/achievements/:id", achievementService.DeleteHandler)
	api.Get("/v1/achievements/me", achievementService.GetMyAchievements)
	api.Get("/v1/achievements/supervised", achievementService.GetSupervisedAchievements)
	api.Post("/v1/achievements/:id/verify", achievementService.Verify)
	api.Post("/v1/achievements/:id/reject", achievementService.Reject)
}
