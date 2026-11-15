package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Reverseflash45/sipinjam-lab/app/service"
	"github.com/Reverseflash45/sipinjam-lab/helper"
)

// Dependencies akan bertambah seiring modul auth, ruang, alat, dan
// peminjaman digabungkan.
type Dependencies struct {
	Pool        *pgxpool.Pool
	AuthService service.AuthService
	JWTSecret   string
}

func Register(app *fiber.App, d Dependencies) {
	api := app.Group("/api/v1")
	api.Get("/health", health(d.Pool))

	// mendaftarkan seluruh endpoint auth
	RegisterAuthRoutes(api, d.AuthService, d.JWTSecret)
}

func health(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			return helper.NewError(fiber.StatusServiceUnavailable, "SERVICE_UNAVAILABLE",
				"PostgreSQL tidak dapat dihubungi")
		}
		return helper.Success(c, fiber.StatusOK, "server dan PostgreSQL berjalan", nil)
	}
}