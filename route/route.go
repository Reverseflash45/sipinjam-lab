package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Reverseflash45/sipinjam-lab/app/model"
	"github.com/Reverseflash45/sipinjam-lab/app/service"
	"github.com/Reverseflash45/sipinjam-lab/helper"
	"github.com/Reverseflash45/sipinjam-lab/middleware"
)

type Dependencies struct {
	Pool           *pgxpool.Pool
	AuthService    service.AuthService
	JWTSecret      string
	RoomService    *service.RoomService
	ItemService    *service.ItemService
	BookingService *service.BookingService
}

// Register memetakan URL ke method service. Berkas ini sekaligus menjadi
// peta hak akses: siapa boleh memanggil endpoint mana.
func Register(app *fiber.App, d Dependencies) {
	api := app.Group("/api/v1")
	api.Get("/health", health(d.Pool))

	// mendaftarkan seluruh endpoint auth
	RegisterAuthRoutes(api, d.AuthService, d.JWTSecret)

	// Semua route di bawah ini wajib login.
	private := api.Group("", middleware.RequireAuth(d.JWTSecret))
	staff := middleware.RequireRole(string(model.RoleLaboran), string(model.RoleAdmin))
	adminOnly := middleware.RequireRole(string(model.RoleAdmin))
	borrower := middleware.RequireRole(string(model.RoleMahasiswa), string(model.RoleDosen))

	private.Get("/rooms", d.RoomService.List)
	private.Get("/rooms/:id", d.RoomService.Get)
	private.Get("/rooms/:id/schedule", d.RoomService.Schedule)
	private.Post("/rooms", adminOnly, d.RoomService.Create)
	private.Put("/rooms/:id", adminOnly, d.RoomService.Update)

	private.Get("/items", d.ItemService.List)
	private.Get("/items/:id", d.ItemService.Get)
	private.Post("/items", staff, d.ItemService.Create)
	private.Put("/items/:id", staff, d.ItemService.Update)

	// "/mine" didaftarkan sebelum "/:id" supaya tidak dianggap id.
	private.Post("/bookings", borrower, d.BookingService.Create)
	private.Get("/bookings/mine", d.BookingService.Mine)
	private.Get("/bookings", staff, d.BookingService.List)
	private.Get("/bookings/:id", d.BookingService.Get)
	private.Patch("/bookings/:id/approve", staff, d.BookingService.Approve)
	private.Patch("/bookings/:id/reject", staff, d.BookingService.Reject)
	private.Patch("/bookings/:id/return", staff, d.BookingService.Return)
	private.Patch("/bookings/:id/cancel", borrower, d.BookingService.Cancel)
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
