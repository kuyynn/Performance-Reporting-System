package middleware

import (
	"strings"
	"uas/utils"

	"github.com/gofiber/fiber/v2"
)

func AuthRequired(secret []byte) fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth == "" { return c.Status(401).JSON(fiber.Map{"error":"missing token"}) }
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 { return c.Status(401).JSON(fiber.Map{"error":"bad auth header"}) }
		claims, err := utils.ParseToken(parts[1], secret)
		if err != nil { return c.Status(401).JSON(fiber.Map{"error":"invalid token"}) }
		c.Locals("claims", claims)
		return c.Next()
	}
}
