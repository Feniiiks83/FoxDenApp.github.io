package http

import (
	"foxDenApp/internal/services"
	userController "foxDenApp/internal/transport/http/controllers/user"

	"github.com/gofiber/fiber/v2"
)

type HTTP struct {
	app            *fiber.App
	userController *userController.UserController
}

func Init(services *services.Services, app *fiber.App) *HTTP {
	return &HTTP{
		app:            app,
		userController: userController.Init(services.UserService),
	}
}

func (http *HTTP) Start() {
	http.userController.Start("user", http.app)
}
