package middleware

import (
	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"github.com/Reverseflash45/sipinjam-lab/helper"
)

func RequireAuth(secret string) fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(secret)},
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return helper.Unauthorized("token tidak valid atau sudah kedaluwarsa")
		},
		SuccessHandler: func(c *fiber.Ctx) error {
			user, ok := c.Locals("user").(*jwt.Token)
			if !ok {
				return helper.Unauthorized("sesi tidak valid")
			}

			claims, ok := user.Claims.(jwt.MapClaims)
			if !ok {
				return helper.Unauthorized("sesi tidak valid")
			}

			if sub, ok := claims["sub"].(float64); ok {
				c.Locals("user_id", int(sub))
			}
			if role, ok := claims["role"].(string); ok {
				c.Locals("role", role)
			}

			return c.Next()
		},
	})
}

func RequireRole(roles ...string) fiber.Handler {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}
	return func(c *fiber.Ctx) error {
		role, _ := c.Locals("role").(string)
		if !allowed[role] {
			return helper.Forbidden("anda tidak memiliki akses ke resource ini")
		}
		return c.Next()
	}
}