package db

import (
	"sync"

	"github.com/google/uuid"
	"github.com/mikeziminio/go-loyalty-system/internal/model"
)

type Storage struct {
	usersByToken sync.Map
	usersByLogin sync.Map
}

func NewStorage() *Storage {
	return &Storage{}
}

func (s *Storage) Register(login string, password string) (*model.User, error) {
	if _, ok := s.usersByLogin.Load(login); ok {
		return nil, model.ErrUserAlreadyExists
	}
	uid, _ := uuid.NewUUID()
	token := uid.String()
	user := model.User{
		Login:    login,
		Password: password,
		Token:    token,
	}
	s.usersByToken.Store(token, &user)
	s.usersByLogin.Store(login, &user)
	return &user, nil
}

func (s *Storage) AuthByLogin(login string, password string) (*model.User, error) {
	u, ok := s.usersByLogin.Load(login)
	if !ok {
		return nil, model.ErrUserNotFound
	}
	user := u.(*model.User)
	if user.Password != password {
		return nil, model.ErrWrongPassword
	}
	return user, nil
}

func (s *Storage) AuthByToken(token string) (*model.User, error) {
	u, ok := s.usersByToken.Load(token)
	if !ok {
		return nil, model.ErrTokenNotFound
	}
	user := u.(*model.User)
	return user, nil
}
