package db

import "time"

type user struct {
	ID        int     `db:"id"`
	Login     string  `db:"login"`
	Password  string  `db:"password"`
	Token     string  `db:"token"`
	Balance   float64 `db:"balance"`
	Withdrawn int     `db:"withdrawn"`
}

type withdrawal struct {
	ID          int       `db:"id"`
	Sum         int       `db:"sum"`
	OrderID     string    `db:"order_id"`
	UserID      int       `db:"user_id"`
	ProcessedAt time.Time `db:"processed_at"`
}
