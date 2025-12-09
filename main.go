package main

import (
	"log"

	"uas/database"
	"uas/app/repository"
	"uas/app/service"
	"uas/config"
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

	// init repos & services
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	authService := service.NewAuthService(userRepo)

	achRepo := repository.NewAchievementRepository(db)
	achService := service.NewAchievementService(achRepo, mongoClient)

	app := fiber.New()

	// routes
	routes.SetupRoutes(
		app,
		userService,
		authService,
		achService,
		middleware.AuthRequired([]byte(cfg.JWTSecret)),
	)

	log.Fatal(app.Listen(":" + cfg.AppPort))
}
