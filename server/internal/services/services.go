package services

import (
	"foxDenApp/internal/config"
	userService "foxDenApp/internal/services/handlers/user"
	"foxDenApp/internal/storage/repositories"
)

type Services struct {
	UserService *userService.UserService
}

func Init(repos *repositories.Repositories, cfg *config.Config) *Services {
	return &Services{
		UserService: userService.Init(repos.UserRepository, cfg),
	}
}
