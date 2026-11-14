package helper

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
)

// ParamID membaca :id sebagai bilangan bulat positif.
func ParamID(c *fiber.Ctx) (int, bool) {
	id, err := c.ParamsInt("id")
	return id, err == nil && id > 0
}

// RequestContext memberi batas waktu 5 detik untuk satu request, supaya
// query yang macet tidak menahan koneksi selamanya.
func RequestContext(c *fiber.Ctx) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.UserContext(), 5*time.Second)
}
