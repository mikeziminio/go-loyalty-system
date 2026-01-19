package db

import (
	"context"
	stdlog "log"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"

	"github.com/mikeziminio/go-loyalty-system/internal/model"
)

var (
	log         *zap.Logger
	pgContainer *postgres.PostgresContainer
	db          *DB
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	var err error
	log, err = zap.NewDevelopment()
	if err != nil {
		stdlog.Fatalf("failed to init logger: %v", err)
	}
	pgContainer = getContainer(ctx)
	connURL, err := pgContainer.ConnectionString(ctx)
	if err != nil {
		log.Fatal("failed to fetch connection URL", zap.Error(err))
	}
	db = getDB(ctx, connURL)

	code := m.Run()

	db.Close()
	pgContainer.Terminate(ctx)

	os.Exit(code)
}

func getContainer(ctx context.Context) *postgres.PostgresContainer {
	postgresContainer, err := postgres.Run(
		ctx, "postgres:18.0-alpine",
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
			wait.ForListeningPort("5432/tcp"),
		),
	)
	if err != nil {
		log.Fatal("failed to init postgres container", zap.Error(err))
	}
	return postgresContainer
}

func getDB(ctx context.Context, connURL string) *DB {
	// add ?sslmode=disable to the connection string
	u, err := url.Parse(connURL)
	if err != nil {
		log.Fatal("failed to parse connection URL", zap.Error(err))
	}
	q := u.Query()
	q.Set("sslmode", "disable")
	u.RawQuery = q.Encode()
	connURL = u.String()

	log.Info("postgres", zap.String("connURL", connURL))

	database, err := NewDB(ctx, connURL, 2, 5, log)
	if err != nil {
		log.Fatal("failed to init DB", zap.Error(err))
	}
	return database
}

func fillUsers(t *testing.T) {
	ctx := t.Context()
	tag, err := db.pool.Exec(ctx, `
		INSERT INTO users(id, login, password, token, balance, withdrawn) VALUES
		(1, 'mike', 'mikepass', 'miketoken', 0, 0),
		(2, 'eugene', 'eugenepass', 'eugenetoken', 10.80, 20),
		(3, 'alyona', 'alyonapass', 'alyonatoken', 0, 40.55)
	`)
	require.NoError(t, err)
	require.True(t, tag.Insert())
	require.Equal(t, int64(3), tag.RowsAffected())
}

func deleteUsers(t *testing.T) {
	ctx := t.Context()
	tag, err := db.pool.Exec(ctx, `DELETE FROM users`)
	require.NoError(t, err)
	require.True(t, tag.Delete())
}

func TestFetchUserByLogin(t *testing.T) {
	ctx := t.Context()
	fillUsers(t)
	defer deleteUsers(t)

	var testCases = []struct {
		name    string
		login   string
		expUser *model.User
		expErr  error
	}{
		{
			name:  "success",
			login: "mike",
			expUser: &model.User{
				ID:        1,
				Login:     "mike",
				Password:  "mikepass",
				Token:     "miketoken",
				Balance:   0,
				Withdrawn: 0,
			},
		},
		{
			name:  "success with balance",
			login: "eugene",
			expUser: &model.User{
				ID:        2,
				Login:     "eugene",
				Password:  "eugenepass",
				Token:     "eugenetoken",
				Balance:   10.80,
				Withdrawn: 20,
			},
		},
		{
			name:   "user not found",
			login:  "nonexistentlogin",
			expErr: model.ErrUserNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			user, err := db.fetchUserByLogin(ctx, tc.login)
			assert.Equal(t, tc.expUser, user)
			assert.Equal(t, tc.expErr, err)
		})
	}
}

func TestFetchUserByToken(t *testing.T) {
	ctx := t.Context()
	fillUsers(t)
	defer deleteUsers(t)

	var testCases = []struct {
		name    string
		token   string
		expUser *model.User
		expErr  error
	}{
		{
			name:  "success",
			token: "miketoken",
			expUser: &model.User{
				ID:        1,
				Login:     "mike",
				Password:  "mikepass",
				Token:     "miketoken",
				Balance:   0,
				Withdrawn: 0,
			},
		},
		{
			name:  "success with balance",
			token: "eugenetoken",
			expUser: &model.User{
				ID:        2,
				Login:     "eugene",
				Password:  "eugenepass",
				Token:     "eugenetoken",
				Balance:   10.80,
				Withdrawn: 20,
			},
		},
		{
			name:   "token not found",
			token:  "nonexistenttoken",
			expErr: model.ErrTokenNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			user, err := db.fetchUserByToken(ctx, tc.token)
			assert.Equal(t, tc.expUser, user)
			assert.Equal(t, tc.expErr, err)
		})
	}
}
