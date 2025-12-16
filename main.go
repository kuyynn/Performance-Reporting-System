package main

import (
	"log"

	"uas/app/repository"
	"uas/app/service"
	"uas/config"
	"uas/database"
	"uas/middleware"
	routes "uas/route"

	"github.com/gofiber/fiber/v2"
)

func main() {
	cfg := config.LoadConfig()

	log.Println("DSN:", cfg.PostgresDSN)

	// CONNECT POSTGRES
	db, err := database.ConnectPostgres(cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("pg connect: %v", err)
	}
	defer db.Close()

	// CONNECT MONGO
	mongoClient, err := database.ConnectMongo(cfg.MongoURI)
	if err != nil {
		log.Fatalf("mongo connect: %v", err)
	}

	// INIT REPOSITORY
	userRepo := repository.NewUserRepository(db)
	studentRepo := repository.NewStudentRepository(db)
	lecturerRepo := repository.NewLecturerRepository(db)
	tokenRepo := repository.NewRefreshTokenRepository(db)
	achRepo := repository.NewAchievementRepository(db)

	// INIT SERVICE
	userService := service.NewUserService(userRepo)
	authService := service.NewAuthService(userRepo, tokenRepo)
	achService := service.NewAchievementService(achRepo, mongoClient)
	adminService := service.NewAdminService(userRepo, studentRepo, lecturerRepo)

	app := fiber.New()

	// ROUTES
	routes.SetupRoutes(
		app,
		userService,
		authService,
		achService,
		adminService,
		middleware.AuthRequired([]byte(cfg.JWTSecret)),
	)

	log.Fatal(app.Listen(":" + cfg.AppPort))
}
