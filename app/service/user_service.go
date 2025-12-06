package service

import (
	"strconv"
	"uas/app/model"
	"uas/app/repository"

	"github.com/gofiber/fiber/v2"
)

type UserService interface {
	Create(c *fiber.Ctx) error
	Update(c *fiber.Ctx) error
	Delete(c *fiber.Ctx) error
	FindById(c *fiber.Ctx) error
	FindAll(c *fiber.Ctx) error
	Logout(c *fiber.Ctx) error
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) Create(c *fiber.Ctx) error {
	var body model.UserCreateRequest
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid_request"})
	}

	_, err := s.repo.Create(c.Context(), body)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "user_created"})
}

func (s *userService) Update(c *fiber.Ctx) error {
	var body model.UserUpdateRequest
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid_request"})
	}

	body.ID, _ = strconv.ParseInt(c.Params("id"), 10, 64)

	err := s.repo.Update(c.Context(), body)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "user_updated"})
}

func (s *userService) Delete(c *fiber.Ctx) error {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)

	err := s.repo.Delete(c.Context(), id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "user_deleted"})
}

func (s *userService) FindById(c *fiber.Ctx) error {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)

	data, err := s.repo.FindById(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "not_found"})
	}

	return c.JSON(data)
}

func (s *userService) FindAll(c *fiber.Ctx) error {
	data, err := s.repo.FindAll(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(data)
}

func (s *userService) Logout(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "logout_success"})
}
