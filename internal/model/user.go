package model

import "errors"

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
}
