package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/mikeziminio/go-loyalty-system/internal/model"
)

type DB struct {
	pool *pgxpool.Pool
	log  *zap.Logger
}

func NewDB(ctx context.Context, connURL string, minConns int, maxConns int, log *zap.Logger) (*DB, error) {
	poolConf, err := pgxpool.ParseConfig(connURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	poolConf.MinConns = int32(minConns)
	poolConf.MaxConns = int32(maxConns)
	pool, err := pgxpool.NewWithConfig(ctx, poolConf)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}
	return &DB{
		pool: pool,
		log:  log,
	}, nil
}

func (db *DB) Begin(ctx context.Context) (pgx.Tx, error) {
	return db.pool.Begin(ctx)
}

func (db *DB) Close() {
	db.pool.Close()
}

func (db *DB) fetchUserByLogin(ctx context.Context, login string) (*model.User, error) {
	q := `SELECT * FROM users WHERE login = $1`
	rows, err := db.pool.Query(ctx, q, login)
	if err != nil {
		return nil, fmt.Errorf("failed to do query: %w", err)
	}
	u, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[user])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrUserNotFound
		}
		return nil, fmt.Errorf("unexpected error: %w", err)
	}
	return &model.User{
		ID:        u.ID,
		Login:     u.Login,
		Password:  u.Password,
		Token:     u.Token,
		Balance:   u.Balance,
		Withdrawn: u.Withdrawn,
	}, nil
}

func (db *DB) fetchUserByID(ctx context.Context, id int) (*model.User, error) {
	q := `SELECT * FROM users WHERE id = $1`
	rows, err := db.pool.Query(ctx, q, id)
	if err != nil {
		return nil, fmt.Errorf("failed to do query: %w", err)
	}
	u, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[user])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrUserNotFound
		}
		return nil, fmt.Errorf("unexpected error: %w", err)
	}
	return &model.User{
		ID:        u.ID,
		Login:     u.Login,
		Password:  u.Password,
		Token:     u.Token,
		Balance:   u.Balance,
		Withdrawn: u.Withdrawn,
	}, nil
}

func (db *DB) fetchUserByToken(ctx context.Context, token string) (*model.User, error) {
	q := `SELECT * FROM users WHERE token = $1`
	rows, err := db.pool.Query(ctx, q, token)
	if err != nil {
		return nil, fmt.Errorf("failed to do query: %w", err)
	}
	u, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[user])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrTokenNotFound
		}
		return nil, fmt.Errorf("unexpected error: %w", err)
	}
	return &model.User{
		ID:        u.ID,
		Login:     u.Login,
		Password:  u.Password,
		Token:     u.Token,
		Balance:   u.Balance,
		Withdrawn: u.Withdrawn,
	}, nil
}

func (db *DB) Register(ctx context.Context, login string, password string) (*model.User, error) {
	q := `SELECT COUNT(id) FROM users WHERE login = $1`
	row := db.pool.QueryRow(ctx, q, login)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("unexpected error: %w", err) // model.ErrUserAlreadyExists
	}
	if count != 0 {
		return nil, model.ErrUserAlreadyExists
	}

	uid, _ := uuid.NewUUID()
	token := uid.String()

	q = `
		INSERT INTO users(login, password, token)
		VALUES ($1, $2, $3)
		RETURNING *
		`
	rows, err := db.pool.Query(ctx, q, login, password, token)
	if err != nil {
		return nil, fmt.Errorf("failed to do query: %w", err)
	}
	u, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[user])
	if err != nil {
		return nil, fmt.Errorf("failed to collect one row: %w", err)
	}

	return &model.User{
		ID:       u.ID,
		Login:    u.Login,
		Password: u.Password,
		Token:    u.Token,
	}, nil
}

func (db *DB) AuthByLogin(ctx context.Context, login string, password string) (*model.User, error) {
	u, err := db.fetchUserByLogin(ctx, login)
	if err != nil {
		return nil, err
	}
	if password != u.Password {
		return nil, model.ErrWrongPassword
	}

	return u, nil
}

func (db *DB) AuthByToken(ctx context.Context, token string) (*model.User, error) {
	u, err := db.fetchUserByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (db *DB) AddWithdrawal(ctx context.Context, userID int, orderID string, sum int) error {
	u, err := db.fetchUserByID(ctx, userID)
	if err != nil {
		return nil
	}
	if u.Balance < float64(sum) {
		return model.ErrInsufficientFunds
	}

	q := `
		UPDATE users
		SET balance = balance - $1, withdrawn = withdrawn + $2
		WHERE id = $3
		`
	t, err := db.pool.Exec(ctx, q, float64(sum), sum, userID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	if !t.Update() || t.RowsAffected() != 1 {
		return fmt.Errorf("failed to update user: %s", t.String())
	}

	q = `
		INSERT INTO withdrawals(sum, order_id, user_id, processed_at)
		VALUES ($1, $2, $3, $4)
		`
	t, err = db.pool.Exec(ctx, q, sum, orderID, userID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}
	if !t.Insert() || t.RowsAffected() != 1 {
		return fmt.Errorf("failed to insert user: %s", t.String())
	}
	return nil
}

func (db *DB) Withdrawals(ctx context.Context, userID int) ([]model.Withdrawal, error) {
	q := `
		SELECT * FROM withdrawals
		WHERE user_id = $1
		ORDER BY processed_at DESC
		`
	rows, err := db.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to do query: %w", err)
	}
	ws, err := pgx.CollectRows(rows, pgx.RowToStructByName[withdrawal])
	if err != nil {
		return nil, fmt.Errorf("failed to collect one row: %w", err)
	}

	var withdrawals []model.Withdrawal
	for _, w := range ws {
		withdrawals = append(withdrawals, model.Withdrawal{
			OrderID:     w.OrderID,
			Sum:         w.Sum,
			ProcessedAt: w.ProcessedAt,
		})
	}
	return withdrawals, nil
}

func (db *DB) AddOrder(ctx context.Context, userID int, orderID string) error {
	q := `
		INSERT INTO orders(id, status, user_id, updated_at)
		VALUES ($1, $2, $3, $4)
		`
	t, err := db.pool.Exec(ctx, q, orderID, "NEW", userID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}
	if !t.Insert() || t.RowsAffected() != 1 {
		return fmt.Errorf("failed to insert order: %s", t.String())
	}
	return nil
}

func (db *DB) Orders(ctx context.Context, userID int) ([]model.Order, error) {
	q := `
		SELECT * FROM orders
		WHERE user_id = $1
		ORDER BY updated_at DESC
		`
	rows, err := db.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to do query: %w", err)
	}
	os, err := pgx.CollectRows(rows, pgx.RowToStructByName[order])
	if err != nil {
		return nil, fmt.Errorf("failed to collect one row: %w", err)
	}

	var orders []model.Order
	for _, o := range os {
		orders = append(orders, model.Order{
			ID:         o.ID,
			Status:     model.OrderStatus(o.Status),
			Accrual:    o.Accrual,
			UploadedAt: o.UpdatedAt,
		})
	}
	return orders, nil
}
