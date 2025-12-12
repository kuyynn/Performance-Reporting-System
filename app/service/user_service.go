package service

import (
	"uas/app/model"
	"uas/app/repository"

	"github.com/gofiber/fiber/v2"
)

type UserService struct {
	Repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return UserService{Repo: repo}
}

func (s *UserService) Create(c *fiber.Ctx) error {
	var input model.UserCreateRequest

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid_request",
		})
	}

	id, err := s.Repo.Create(c.Context(), input)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"id": id,
	})
}
