package model

import (
	"errors"
	"time"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrTokenNotFound     = errors.New("token not found")
	ErrUserNotFound      = errors.New("user not found")
	ErrWrongPassword     = errors.New("wrong password")
)

type User struct {
	ID       int
	Login    string
	Password string
	Token    string

	Balance   float64
	Withdrawn int
}

type Withdrawal struct {
	OrderID     string    `json:"order"`
	Sum         int       `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}
