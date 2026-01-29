package user

import "github.com/gofiber/fiber/v2"

// Handlers defines the user HTTP handlers interface
type Handlers interface {
	Signup() fiber.Handler
	Login() fiber.Handler
	Logout() fiber.Handler
	GetCurrentUser() fiber.Handler
}
