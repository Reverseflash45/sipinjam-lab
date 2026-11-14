package middleware

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

// Register memasang middleware global. Urutannya penting: recover paling
// luar supaya panic di mana pun tetap menjadi response 500.
func Register(app *fiber.App, logger *slog.Logger, allowedOrigins string) {
	app.Use(recover.New())
	app.Use(requestid.New())
	if allowedOrigins != "" {
		app.Use(cors.New(cors.Config{AllowOrigins: allowedOrigins}))
	}
	app.Use(func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		logger.Info("request",
			"method", c.Method(), "path", c.Path(),
			"status", c.Response().StatusCode(), "ms", time.Since(start).Milliseconds())
		return err
	})
}
