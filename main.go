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
	// LOAD CONFIG
	cfg := config.LoadConfig()

	log.Println("DSN:", cfg.PostgresDSN)

	// =========================
	// CONNECT POSTGRES
	// =========================
	db, err := database.ConnectPostgres(cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("pg connect error: %v", err)
	}
	defer db.Close()

	// =========================
	// CONNECT MONGO
	// =========================
	mongoClient, err := database.ConnectMongo(cfg.MongoURI)
	if err != nil {
		log.Fatalf("mongo connect error: %v", err)
	}

	mongoDB := mongoClient.Database("uas")

	// =========================
	// INIT REPOSITORIES
	// =========================
	userRepo := repository.NewUserRepository(db)
	studentRepo := repository.NewStudentRepository(db)
	lecturerRepo := repository.NewLecturerRepository(db)
	tokenRepo := repository.NewRefreshTokenRepository(db)
	achievementRepo := repository.NewAchievementRepository(db)

	mongoAchievementRepo := repository.NewMongoAchievementRepository(
		mongoDB.Collection("achievements"),
	)

	// =========================
	// INIT SERVICES
	// =========================
	userService := service.NewUserService(userRepo)
	authService := service.NewAuthService(userRepo, tokenRepo)
	achievementService := service.NewAchievementService(achievementRepo, mongoClient)
	adminService := service.NewAdminService(userRepo, studentRepo, lecturerRepo)
	studentService := service.NewStudentService(studentRepo)
	lecturerService := service.NewLecturerService(lecturerRepo)

	reportService := &service.ReportService{
		AchievementRepo:      achievementRepo,
		MongoAchievementRepo: mongoAchievementRepo,
	}

	// =========================
	// INIT APP
	// =========================
	app := fiber.New()

	app.Static("/docs", "./docs")

	// =========================
	// SETUP ROUTES
	// =========================
	routes.SetupRoutes(
		app,
		userService,
		authService,
		achievementService,
		adminService,
		middleware.AuthRequired([]byte(cfg.JWTSecret)),
		studentService,
		lecturerService,
		reportService,
	)

	// START SERVER
	// =========================
	log.Fatal(app.Listen(":" + cfg.AppPort))
}
