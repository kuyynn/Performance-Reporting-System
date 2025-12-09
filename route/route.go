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

	// ====== PUBLIC ======

	// Login tetap di sini (tapi bisa kita refactor nanti)
	app.Post("/login", authService.LoginHandler)

	// ====== PROTECTED ======
	api := app.Group("/api", authMiddleware)

	// USER CRUD
	api.Post("/users", userService.Create)
	api.Get("/users", userService.FindAll)
	api.Get("/users/:id", userService.FindById)
	api.Put("/users/:id", userService.Update)
	api.Delete("/users/:id", userService.Delete)

	api.Post("/logout", userService.Logout)

	// ACHIEVEMENT (CLEAN ROUTER)
	api.Post("/v1/achievements", achievementService.Create)
	api.Post("/v1/achievements/:id/submit", achievementService.Submit)
	api.Get("/v1/achievements/me", achievementService.GetMyAchievements)

}
