package db

import (
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mikeziminio/go-loyalty-system/internal/model"
)

type Storage struct {
	usersByToken        sync.Map
	usersByLogin        sync.Map
	usersByID           sync.Map
	withdrawalsByUserID sync.Map
	ordersByUserID      sync.Map
	ordersByID          sync.Map
}

func NewStorage() *Storage {
	return &Storage{}
}

var userIDCounter int

func (s *Storage) Register(login string, password string) (*model.User, error) {
	if _, ok := s.usersByLogin.Load(login); ok {
		return nil, model.ErrUserAlreadyExists
	}
	uid, _ := uuid.NewUUID()
	token := uid.String()
	userIDCounter++
	user := model.User{
		ID:       userIDCounter,
		Login:    login,
		Password: password,
		Token:    token,
	}
	s.usersByToken.Store(token, &user)
	s.usersByLogin.Store(login, &user)
	s.usersByID.Store(userIDCounter, &user)
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

func (s *Storage) AddWithdrawal(userID int, orderID string, sum int) error {
	u, ok := s.usersByID.Load(userID)
	if !ok {
		return model.ErrUserNotFound
	}
	user := u.(*model.User)

	if user.Balance < float64(sum) {
		return model.ErrInsufficientFunds
	}
	user.Balance -= float64(sum)
	user.Withdrawn += sum

	var ws []model.Withdrawal
	if aws, ok := s.withdrawalsByUserID.Load(userID); ok {
		ws, ok = aws.([]model.Withdrawal)
		if !ok {
			return fmt.Errorf("failed to load value from sync map")
		}
	}
	ws = append(ws, model.Withdrawal{
		OrderID:     orderID,
		Sum:         sum,
		ProcessedAt: time.Now(),
	})
	s.withdrawalsByUserID.Store(userID, ws)
	return nil
}

func (s *Storage) Withdrawals(userID int) ([]model.Withdrawal, error) {
	aws, ok := s.withdrawalsByUserID.Load(userID)
	if !ok {
		return nil, nil
	}
	ws, ok := aws.([]model.Withdrawal)
	if !ok {
		return nil, fmt.Errorf("failed to load value from sync map")
	}
	ws = slices.Clone(ws)
	slices.Reverse(ws)
	return ws, nil
}

func (s *Storage) Orders(userID int) ([]model.Withdrawal, error) {
	aws, ok := s.withdrawalsByUserID.Load(userID)
	if !ok {
		return nil, nil
	}
	ws, ok := aws.([]model.Withdrawal)
	if !ok {
		return nil, fmt.Errorf("failed to load value from sync map")
	}
	ws = slices.Clone(ws)
	slices.Reverse(ws)
	return ws, nil
}
