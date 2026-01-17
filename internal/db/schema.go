package db

import "time"

// todo:
// float64 to decimal
// https://github.com/shopspring/decimal

type user struct {
	ID        int     `db:"id"`
	Login     string  `db:"login"`
	Password  string  `db:"password"`
	Token     string  `db:"token"`
	Balance   float64 `db:"balance"`
	Withdrawn float64 `db:"withdrawn"`
}

type withdrawal struct {
	ID          int       `db:"id"`
	Sum         float64   `db:"sum"`
	OrderID     string    `db:"order_id"`
	UserID      int       `db:"user_id"`
	ProcessedAt time.Time `db:"processed_at"`
}

type order struct {
	ID        string    `db:"id"`
	Status    string    `db:"status"`
	Accrual   float64   `db:"accrual"`
	UserID    int       `db:"user_id"`
	UpdatedAt time.Time `db:"updated_at"`
}
