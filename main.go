package main

import (
	// "database/sql"
	"log"
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

	db, err := config.ConnectPostgres(cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("pg connect: %v", err)
	}
	defer db.Close()

	// init repos & services
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	authService := service.NewAuthService(userRepo)

	app := fiber.New()

	// routes: combine user & auth
	routes.SetupRoutes(app, userService, authService, middleware.AuthRequired([]byte(cfg.JWTSecret)))

	log.Fatal(app.Listen(":" + cfg.AppPort))
}
