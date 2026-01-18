package test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func Logger(t *testing.T) *zap.Logger {
	t.Helper()
	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	return log
}
