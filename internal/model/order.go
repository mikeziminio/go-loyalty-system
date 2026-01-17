package model

import (
	"errors"
	"time"
)

type OrderStatus string

var (
	StatusRegistered OrderStatus = "REGISTERED"
	StatusProcessed  OrderStatus = "PROCESSED"
	StatusProcessing OrderStatus = "PROCESSING"
	StatusInvalid    OrderStatus = "INVALID"
)

var (
	ErrInsufficientFunds        = errors.New("insufficient funds")
	ErrOrderAlreadyLoaded       = errors.New("order already loaded")
	ErrOrderLoadedByAnotherUser = errors.New("order loaded by another user")
)

type Order struct {
	ID         string
	Status     OrderStatus
	Accrual    float64
	UploadedAt time.Time
}
