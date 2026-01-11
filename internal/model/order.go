package model

import "time"

type OrderStatus string

var (
	StatusRegistered OrderStatus = "REGISTERED"
	StatusProcessed  OrderStatus = "PROCESSED"
	StatusProcessing OrderStatus = "PROCESSING"
	StatusInvalid    OrderStatus = "INVALID"
)

type OrderInfo struct {
	Id         string
	Status     OrderStatus
	Accrual    int
	UploadedAt time.Time
}
