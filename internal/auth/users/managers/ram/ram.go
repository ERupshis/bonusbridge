package ram

import (
	"context"
	"errors"
	"fmt"

	"github.com/erupshis/bonusbridge/internal/auth/users/data"
	"github.com/erupshis/bonusbridge/internal/auth/users/managers"
	"github.com/erupshis/bonusbridge/internal/logger"
)

var usersStorage = []data.User{
	{Login: "u1", Password: "p1", ID: 1, Role: data.RoleAdmin},
	{Login: "user2", Password: "password2", ID: 2, Role: data.RoleUser},
}

type Storage struct {
	users []data.User

	log logger.BaseLogger
}

func Create(baseLogger logger.BaseLogger) managers.BaseUsersManager {
	return &Storage{
		log:   baseLogger,
		users: usersStorage,
	}
}

func (s *Storage) AddUser(_ context.Context, user *data.User) (int64, error) {
	s.users = append(s.users, data.User{ID: int64(len(s.users)), Login: user.Login, Password: user.Password, Role: data.RoleUser})

	userInDB, err := s.getUser(user.Login)
	if err != nil {
		if errors.Is(err, data.ErrUserNotFound) {
			return -1, nil
		}

		return -1, err
	}

	return userInDB.ID, nil
}

func (s *Storage) GetUser(_ context.Context, login string) (*data.User, error) {
	user, err := s.getUser(login)
	if err != nil {
		if errors.Is(err, data.ErrUserNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &user, nil
}

func (s *Storage) GetUserID(_ context.Context, login string) (int64, error) {
	user, err := s.getUser(login)
	if err != nil {
		if errors.Is(err, data.ErrUserNotFound) {
			return -1, nil
		}

		return -1, err
	}

	return user.ID, nil
}

func (s *Storage) GetUserRole(_ context.Context, userID int64) (int, error) {
	user, err := s.getUserByID(userID)
	if err != nil {
		if errors.Is(err, data.ErrUserNotFound) {
			return -1, nil
		}

		return -1, err
	}

	return user.Role, nil
}

func (s *Storage) ValidateUser(_ context.Context, login string, password string) (bool, error) {
	user, err := s.getUser(login)
	if err != nil {
		return false, fmt.Errorf("validate user: %w", err)
	}

	if user.ID == -1 {
		return false, fmt.Errorf("validate user: user not found")
	}

	userPwd, err := s.getUserPassword(login)
	if err != nil {
		return false, fmt.Errorf("validate user: %w", err)
	}

	return password == userPwd, nil
}

func (s *Storage) getUser(login string) (data.User, error) {
	for _, u := range s.users {
		if login == u.Login {
			return u, nil
		}
	}

	return data.User{}, data.ErrUserNotFound
}

func (s *Storage) getUserByID(userID int64) (data.User, error) {
	for idx, u := range s.users {
		if userID == int64(idx) {
			return u, nil
		}
	}

	return data.User{}, data.ErrUserNotFound
}

func (s *Storage) getUserPassword(login string) (string, error) {
	user, err := s.getUser(login)
	if err != nil {
		return "", err
	}

	return user.Password, nil
}
