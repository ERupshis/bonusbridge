package ramusers

import (
	"errors"
	"fmt"

	"github.com/erupshis/bonusbridge/internal/auth/users"
	"github.com/erupshis/bonusbridge/internal/logger"
)

// errNotFound missing user in database.
var errNotFound = fmt.Errorf("user not found")

// User represents a user in our system.
type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`

	id   int
	role int
}

var usersStorage = []User{
	{Login: "u1", Password: "p1", id: 1, role: users.RoleAdmin},
	{Login: "user2", Password: "password2", id: 2, role: users.RoleUser},
}

type Storage struct {
	users []User

	log logger.BaseLogger
}

func Create(baseLogger logger.BaseLogger) users.BaseUsers {
	return &Storage{
		log:   baseLogger,
		users: usersStorage,
	}
}

func (s *Storage) AddUser(login string, password string) (int, error) {
	s.users = append(s.users, User{id: len(s.users), Login: login, Password: password, role: users.RoleUser})

	user, err := s.getUser(login)
	if err != nil {
		if errors.As(err, &errNotFound) {
			return -1, nil
		}

		return -1, err
	}

	return user.id, nil
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

func (s *Storage) GetUserRole(userID int) (int, error) {
	user, err := s.getUserByID(userID)
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
	user, err := s.getUser(login)
	if err != nil {
		return false, fmt.Errorf("validate user: %w", err)
	}

	if user.id == -1 {
		return false, fmt.Errorf("validate user: user not found")
	}

	userPwd, err := s.getUserPassword(login)
	if err != nil {
		return false, fmt.Errorf("validate user: %w", err)
	}

	return password == userPwd, nil
}

func (s *Storage) getUser(login string) (User, error) {
	for _, u := range s.users {
		if login == u.Login {
			return u, nil
		}
	}

	return User{}, errNotFound
}

func (s *Storage) getUserByID(id int) (User, error) {
	for idx, u := range s.users {
		if id == idx {
			return u, nil
		}
	}

	return User{}, errNotFound
}

func (s *Storage) getUserPassword(login string) (string, error) {
	user, err := s.getUser(login)
	if err != nil {
		return "", err
	}

	return user.Password, nil
}
