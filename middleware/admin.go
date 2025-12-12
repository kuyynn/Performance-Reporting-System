package middleware

import (
	"uas/utils"
	"github.com/gofiber/fiber/v2"
)

func AdminOnly(c *fiber.Ctx) error {
	claims := c.Locals("claims")

	if claims == nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	userClaims := claims.(*utils.Claims)

	if userClaims.Role != "admin" {
		return c.Status(403).JSON(fiber.Map{"error": "admin_only"})
	}

	return c.Next()
}
