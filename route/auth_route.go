package route

import (
	"github.com/gofiber/fiber/v2"

	"github.com/Reverseflash45/sipinjam-lab/app/service"
	"github.com/Reverseflash45/sipinjam-lab/middleware"
)

func RegisterAuthRoutes(api fiber.Router, svc service.AuthService, jwtSecret string) {
	auth := api.Group("/auth")
	auth.Post("/register", svc.Register)
	auth.Post("/login", svc.Login)
	auth.Get("/me", middleware.RequireAuth(jwtSecret), svc.Me)
}