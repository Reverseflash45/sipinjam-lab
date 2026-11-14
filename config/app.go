package config

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"github.com/Reverseflash45/sipinjam-lab/app/model"
	"github.com/Reverseflash45/sipinjam-lab/helper"
	"github.com/Reverseflash45/sipinjam-lab/middleware"
	"github.com/Reverseflash45/sipinjam-lab/route"
)

// NewApp merakit aplikasi: instance Fiber, middleware global, route, dan
// satu error handler untuk seluruh aplikasi.
func NewApp(logger *slog.Logger, deps route.Dependencies) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      GetEnv("APP_NAME", "SiPinjam Lab"),
		ErrorHandler: errorHandler(logger),
		BodyLimit:    1 * 1024 * 1024,
	})

	middleware.Register(app, logger, GetEnv("ALLOWED_ORIGINS", ""))
	route.Register(app, deps)

	app.Use(func(c *fiber.Ctx) error { return helper.NotFound("endpoint tidak ditemukan") })
	return app
}

// errorHandler adalah satu-satunya tempat error berubah menjadi response
// HTTP. Service cukup mengembalikan error.
func errorHandler(logger *slog.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		var appErr *helper.AppError
		var fiberErr *fiber.Error
		switch {
		case errors.As(err, &appErr):
		case errors.As(err, &fiberErr):
			appErr = helper.NewError(fiberErr.Code, "HTTP_ERROR", fiberErr.Message)
		default:
			appErr = helper.Internal(err)
		}

		if appErr.Status >= 500 {
			logger.Error("request_failed", "path", c.Path(), "error", appErr.Error())
		} else {
			logger.Warn("request_rejected", "path", c.Path(), "code", appErr.Code)
		}

		return c.Status(appErr.Status).JSON(model.ErrorResponse{
			Success: false, Code: appErr.Code, Message: appErr.Message, Fields: appErr.Fields,
		})
	}
}
