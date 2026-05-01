package UserController

import (
	_ "foxDenApp/internal/dto/user"
	user_dto "foxDenApp/internal/dto/user"
	"foxDenApp/internal/models"
	userService "foxDenApp/internal/services/handlers/user"
	_ "foxDenApp/internal/types"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

type UserController struct {
	userService *userService.UserService
}

func Init(userService *userService.UserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

func (controller *UserController) Start(route string, app *fiber.App) {
	router := app.Group("/" + route)

	router.Get("/allusers", controller.FindAllUsers)
	router.Get("/by-ip", controller.FindUsersByIp)
	router.Get("/by-time", controller.FindUsersByTime)
	router.Get("/:id", controller.FindUserById)

	router.Post("/create", controller.CreateUser)

	router.Delete("/:id", controller.DeleteUser)
}

// CreateUser godoc
//
//	@Summary		Create a user
//	@Description	Create a new user.
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Security ApiKeyAuth
//	@Param 			body body user_dto.CreateUserResponse true "Body"
//	@Success		201	{object}	models.User
//	@Failure		400 {object}	object{message=string,error=string}
//	@Failure		404 {object}	object{message=string,error=string}
//	@Failure		500 {object}	object{message=string,error=string}
//	@Router			/user/create [post]
func (controller *UserController) CreateUser(c *fiber.Ctx) error {
	var body user_dto.CreateUserResponse
	var createduser *models.User

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.ErrBadRequest.Code).JSON(fiber.Map{
			"message": "Couldn't parse body!",
			"error":   err.Error(),
		})
	}

	createduser, err := controller.userService.CreateUser(body.Time, body.Ip)

	if err != nil {
		return c.Status(fiber.ErrInternalServerError.Code).JSON(fiber.Map{
			"message": "Couldn't create user!",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(createduser)
}

// FindUserById godoc
//
//	@Summary		Find user by id
//	@Description	Find user by his id
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param 			id path int true "user's id"
//	@Success		200	{object}	models.User
//	@Failure		400 {object}	object{message=string,error=string}
//	@Failure		404 {object}	object{message=string,error=string}
//	@Failure		500 {object}	object{message=string,error=string}
//	@Router			/user/{id} [get]
func (controller *UserController) FindUserById(c *fiber.Ctx) error {
	id := c.Params("id")
	uintId, err := strconv.ParseUint(id, 10, 64)

	if err != nil {
		return c.Status(fiber.ErrBadRequest.Code).JSON(fiber.Map{
			"message": "Wrong id format!",
			"error":   err.Error(),
		})
	}

	user, err := controller.userService.FindUserById(uintId)

	if err != nil {
		return c.Status(fiber.ErrInternalServerError.Code).JSON(fiber.Map{
			"message": "Couldn't find user!",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(user)
}

// FindUsersByIp godoc
//
//	@Summary		Find user by ip
//	@Description	Find user by his ip
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param 			ip query string true "user's ip"
//	@Success		200	{object}	models.Users
//	@Failure		400 {object}	object{message=string,error=string}
//	@Failure		404 {object}	object{message=string,error=string}
//	@Failure		500 {object}	object{message=string,error=string}
//	@Router			/user/by-ip [get]
func (controller *UserController) FindUsersByIp(c *fiber.Ctx) error {
	ip := c.Query("ip")

	if len(ip) == 0 {
		return c.Status(fiber.ErrBadRequest.Code).JSON(fiber.Map{
			"message": "Wrong ip format!",
		})
	}

	users, err := controller.userService.FindUsersByIp(ip)

	if err != nil {
		return c.Status(fiber.ErrInternalServerError.Code).JSON(fiber.Map{
			"message": "Couldn't find user!",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(users)
}

// FindAllUsers godoc
//
//	@Summary		Find users by ip
//	@Description	Retrieve a list of all users from the database
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	models.Users
//	@Failure		400 {object}	object{message=string,error=string}
//	@Failure		404 {object}	object{message=string,error=string}
//	@Failure		500 {object}	object{message=string,error=string}
//	@Router			/user/allusers [get]
func (controller *UserController) FindAllUsers(c *fiber.Ctx) error {
	users, err := controller.userService.FindAllUsers()

	if err != nil {
		return c.Status(fiber.ErrInternalServerError.Code).JSON(fiber.Map{
			"message": "Couldn't find user!",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(users)
}

// FindUsersByTime godoc
//
//	@Summary		Find users by time
//	@Description	Get users created between timefrom and timebefore
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param 			timefrom   query string true "Start time (RFC3339, e.g. 2024-05-01T15:00:00Z)"
//	@Param 			timebefore query string true "End time (RFC3339, e.g. 2024-05-12T15:00:00Z)"
//	@Success		200	{object}	models.Users
//	@Failure		400 {object}	object{message=string,error=string}
//	@Failure		500 {object}	object{message=string,error=string}
//	@Router			/user/by-time [get]
func (controller *UserController) FindUsersByTime(c *fiber.Ctx) error {
	timeFromStr := c.Query("timefrom")
	timeBeforeStr := c.Query("timebefore")

	timeFrom, err := time.Parse(time.RFC3339, timeFromStr)

	if err != nil {
		return c.Status(fiber.ErrBadRequest.Code).JSON(fiber.Map{
			"message": "Wrong timeFrom format!",
			"error":   err.Error(),
		})
	}

	timeBefore, err := time.Parse(time.RFC3339, timeBeforeStr)

	if err != nil {
		return c.Status(fiber.ErrBadRequest.Code).JSON(fiber.Map{
			"message": "Wrong timeBefore format!",
			"error":   err.Error(),
		})
	}

	users, err := controller.userService.FindUsersByTime(timeFrom, timeBefore)

	if err != nil {
		return c.Status(fiber.ErrInternalServerError.Code).JSON(fiber.Map{
			"message": "Couldn't find user!",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(users)
}

// DeleteUser godoc
//
//	@Summary		Delete user
//	@Description	Delete user by id
//	@Tags			user
//	@Param 			id  path int true "user's id"
//	@Accept			json
//	@Produce		json
//	@Security ApiKeyAuth
//	@Success		200	{object}	object{status=int}
//	@Failure		400 {object}	object{message=string,error=string}
//	@Failure		404 {object}	object{message=string,error=string}
//	@Failure		500 {object}	object{message=string,error=string}
//	@Router			/user/{id} [delete]
func (controller *UserController) DeleteUser(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)

	if err != nil {
		return c.Status(fiber.ErrBadRequest.Code).JSON(fiber.Map{
			"message": "Incorrect param (id) type!",
			"error":   err.Error(),
		})
	}

	err = controller.userService.DeleteUser(id)

	if err != nil {
		return c.Status(fiber.ErrInternalServerError.Code).JSON(fiber.Map{
			"message": "Couldn't delete user!",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": fiber.StatusOK,
	})
}
