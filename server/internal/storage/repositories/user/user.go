package UserRepo

import (
	user_dto "foxDenApp/internal/dto/user"
	"foxDenApp/internal/models"
	"foxDenApp/internal/storage"
	"time"
)

type UserRepo struct {
	db *storage.Storage
}

func InitUserRepository(db *storage.Storage) *UserRepo {
	UserRepo := UserRepo{
		db: db,
	}

	return &UserRepo
}

func (repo *UserRepo) CreateUser(createUser *user_dto.CreateUserResponse) (*models.User, error) {
	var result models.User

	err := repo.db.Db.QueryRow(
		`
		INSERT INTO users (time, ip) 
		VALUES ($1, $2) 
		RETURNING id, time, ip
		`,
		createUser.Time,
		createUser.Ip,
	).Scan(&result.Id, &result.Time, &result.Ip)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (repo *UserRepo) FindUserById(id uint64) (*models.User, error) {
	var result models.User

	err := repo.db.Db.QueryRow(
		`
		SELECT id, time, ip 
		FROM users 
		WHERE id = $1
		`,
		id,
	).Scan(&result.Id, &result.Time, &result.Ip)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (repo *UserRepo) FindUsersByIp(ip string) (*models.Users, error) {
	var result []models.User

	rows, err := repo.db.Db.Query(
		`
		SELECT id, time, ip 
		FROM users 
		WHERE ip = $1
		`,
		ip,
	)

	var user models.User
	for rows.Next() {
		err := rows.Scan(&user.Id, &user.Time, &user.Ip)

		if err != nil {
			return nil, err
		}
		result = append(result, user)
	}

	if err != nil {
		return nil, err
	}

	return &models.Users{
		Users: result,
		Count: uint64(len(result)),
	}, nil
}

func (repo *UserRepo) FindAllUsers() (*models.Users, error) {
	var result []models.User

	rows, err := repo.db.Db.Query(
		`
		SELECT * FROM users
		`,
	)

	if err != nil {
		return nil, err
	}

	var user models.User
	for rows.Next() {
		err := rows.Scan(&user.Id, &user.Time, &user.Ip)

		if err != nil {
			return nil, err
		}
		result = append(result, user)
	}

	return &models.Users{
		Users: result,
		Count: uint64(len(result)),
	}, nil
}

func (repo *UserRepo) FindUsersByTime(timeFrom time.Time, timeBefore time.Time) (*models.Users, error) {
	var result []models.User

	rows, err := repo.db.Db.Query(
		`
		SELECT * FROM users WHERE time BETWEEN $1 AND $2;

		`,
		timeFrom,
		timeBefore,
	)

	if err != nil {
		return nil, err
	}

	var user models.User
	for rows.Next() {
		err := rows.Scan(&user)

		if err != nil {
			return nil, err
		}
		result = append(result, user)
	}

	return &models.Users{
		Users: result,
		Count: uint64(len(result)),
	}, nil
}

func (repo *UserRepo) DeleteUserData(id uint64) error {
	err := repo.db.Db.QueryRow(
		`
		DELETE FROM users 
		WHERE id = $1
		`,
		id,
	).Err()

	if err != nil {
		return err
	}

	return nil
}
