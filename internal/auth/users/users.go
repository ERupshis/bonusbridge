package users

import (
	"errors"
	"fmt"

	"github.com/erupshis/bonusbridge/internal/logger"
)

const (
	RoleUser = iota
	RoleAdmin
)

var errNotFound = fmt.Errorf("user not found")

// User represents a user in our system.
type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`

	id   int
	role int
}

var users = []User{
	{Login: "u1", Password: "p1", id: 1, role: RoleAdmin},
	{Login: "user2", Password: "password2", id: 2, role: RoleUser},
}

type Storage struct {
	users []User

	log logger.BaseLogger
}

func Create(baseLogger logger.BaseLogger) Storage {
	return Storage{
		log:   baseLogger,
		users: users,
	}
}

func (s *Storage) AddUser(login string, password string) (int, error) {
	s.users = append(s.users, User{id: len(s.users), Login: login, Password: password, role: RoleUser})

	user, err := s.getUser(login)
	if err != nil {
		if errors.As(err, &errNotFound) {
			return -1, nil
		}

		return -1, err
	}

	return user.id, nil
}

func (s *Storage) getUser(login string) (User, error) {
	for _, u := range s.users {
		if login == u.Login {
			return u, nil
		}
	}

	return User{}, errNotFound
}

func (s *Storage) GetUserId(login string) (int, error) {
	user, err := s.getUser(login)
	if err != nil {
		if errors.As(err, &errNotFound) {
			return -1, nil
		}

		return -1, err
	}

	return user.id, nil
}

func (s *Storage) GetUserRole(login string) (int, error) {
	user, err := s.getUser(login)
	if err != nil {
		if errors.As(err, &errNotFound) {
			return -1, nil
		}

		return -1, err
	}

	return user.role, nil
}

func (s *Storage) ValidateUser(login string, password string) (bool, error) {
	//TODO: need implement hash with asymmetric keys
	exists, err := s.HasUser(login)
	if err != nil {
		return false, fmt.Errorf("validate user: %w", err)
	}

	if !exists {
		return false, fmt.Errorf("validate user: user not found")
	}

	userPwd, err := s.getUserPassword(login)
	if err != nil {
		return false, fmt.Errorf("validate user: %w", err)
	}

	return password == userPwd, nil
}

func (s *Storage) HasUser(name string) (bool, error) {
	user, err := s.getUser(name)
	if err != nil {
		if errors.As(err, &errNotFound) {
			return false, nil
		}

		return false, err
	}

	return user.id != -1, nil
}

func (s *Storage) getUserPassword(login string) (string, error) {
	user, err := s.getUser(login)
	if err != nil {
		return "", err
	}

	return user.Password, nil
}
