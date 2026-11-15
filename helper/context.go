package helper

import (
	"github.com/gofiber/fiber/v2"

	"github.com/Reverseflash45/sipinjam-lab/app/model"
)

// GetAuthUser membaca identitas user yang sudah diset middleware RequireAuth
// ke Fiber Locals. Mengembalikan (AuthUser, true) bila lengkap,
// (AuthUser{}, false) bila tidak ada / tidak valid.
func GetAuthUser(c *fiber.Ctx) (model.AuthUser, bool) {
	id, ok := c.Locals("user_id").(int)
	if !ok || id <= 0 {
		return model.AuthUser{}, false
	}

	role, _ := c.Locals("role").(string)
	return model.AuthUser{UserID: id, Role: model.Role(role)}, true
}

// MustAuthUser sama seperti GetAuthUser tetapi mengembalikan AppError 401
// bila identitas tidak tersedia.
func MustAuthUser(c *fiber.Ctx) (model.AuthUser, error) {
	u, ok := GetAuthUser(c)
	if !ok {
		return model.AuthUser{}, Unauthorized("sesi tidak valid")
	}
	return u, nil
}

// CurrentUser alias MustAuthUser, disediakan untuk kompatibilitas
// dengan kode service yang memakai helper.CurrentUser(c).
func CurrentUser(c *fiber.Ctx) (model.AuthUser, error) {
	return MustAuthUser(c)
}