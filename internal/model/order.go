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
	ErrInsufficientFunds = errors.New("insufficient funds")
)

type Order struct {
	Id         string
	Status     OrderStatus
	Accrual    int
	UploadedAt time.Time
}
