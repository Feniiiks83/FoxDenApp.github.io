package user_service

import (
	"foxDenApp/internal/config"
	user_dto "foxDenApp/internal/dto/user"
	"foxDenApp/internal/models"
	userRepo "foxDenApp/internal/storage/repositories/user"
	"time"
)

type UserService struct {
	repo *userRepo.UserRepo
	cfg  *config.Config
}

func Init(userRepo *userRepo.UserRepo, cfg *config.Config) *UserService {
	return &UserService{
		repo: userRepo,
		cfg:  cfg,
	}
}

func (s *UserService) FindUserById(id uint64) (*models.User, error) {
	debtor, err := s.repo.FindUserById(id)

	if err != nil {
		return nil, err
	}

	return debtor, nil
}

func (s *UserService) CreateUser(time time.Time, ip string) (*models.User, error) {
	createUser := user_dto.CreateUserResponse{
		Time: time,
		Ip:   ip,
	}

	createdUser, err := s.repo.CreateUser(&createUser)

	if err != nil {
		return nil, err
	}

	return createdUser, nil
}

func (s *UserService) FindUsersByIp(ip string) (*models.Users, error) {
	users, err := s.repo.FindUsersByIp(ip)

	if err != nil {
		return nil, err
	}

	return users, err
}

func (s *UserService) FindUsersByTime(timeFrom time.Time, timeBefore time.Time) (*models.Users, error) {
	users, err := s.repo.FindUsersByTime(timeFrom, timeBefore)

	if err != nil {
		return nil, err
	}

	return users, err
}

func (s *UserService) FindAllUsers() (*models.Users, error) {
	users, err := s.repo.FindAllUsers()

	if err != nil {
		return nil, err
	}

	return users, err
}

func (s *UserService) DeleteUser(id uint64) error {
	err := s.repo.DeleteUserData(id)

	if err != nil {
		return err
	}

	return nil
}
