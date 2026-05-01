package repositories

import (
	"foxDenApp/internal/storage"
	userRepo "foxDenApp/internal/storage/repositories/user"
)

type Repositories struct {
	UserRepository *userRepo.UserRepo
}

func InitRepositories(db *storage.Storage) *Repositories {
	UserRepository := userRepo.InitUserRepository(db)

	Repositories := Repositories{
		UserRepository: UserRepository,
	}

	return &Repositories
}
