package middleware

import (
	"github.com/Hari-Krishna-Moorthy/orders/internals/platform/config"
	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v3"
	"github.com/golang-jwt/jwt/v5"
)

func JWT() fiber.Handler {
	cfg := config.Get()
	return jwtware.New(jwtware.Config{
		SigningKey:    []byte(cfg.JWT.AccessSecret),
		SigningMethod: jwt.SigningMethodHS256.Name,
		ContextKey:    "user",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		},
	})
}
