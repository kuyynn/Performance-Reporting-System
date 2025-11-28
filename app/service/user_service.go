package service

import (
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
	Login(c *fiber.Ctx) error
}

type UserRepositoryy struct {
	repo repository.UserRepositoryy
}

func (u UserRepositoryy) Create(c *fiber.Ctx) error {
	//TODO implement me
	panic("implement me")
}

func (u UserRepositoryy) Update(c *fiber.Ctx) error {
	//TODO implement me
	panic("implement me")
}

func (u UserRepositoryy) Delete(c *fiber.Ctx) error {
	//TODO implement me
	panic("implement me")
}

func (u UserRepositoryy) FindById(c *fiber.Ctx) error {
	//TODO implement me
	panic("implement me")
}

func (u UserRepositoryy) FindAll(c *fiber.Ctx) error {
	//TODO implement me
	panic("implement me")
}

func (u UserRepositoryy) Logout(c *fiber.Ctx) error {
	//TODO implement me
	panic("implement me")
}

func (u UserRepositoryy) Login(c *fiber.Ctx) error {
	panic("implement me")
}