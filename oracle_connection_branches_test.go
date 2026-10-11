package oracle_storage

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestOracleConnectionManager(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	mock.ExpectClose()
	db2, mock2, err := sqlmock.New()
	assert.NoError(t, err)
	mock2.ExpectClose()

	manager := NewOracleConnectionManagerFromDSN("oracle://u:p@h:1/s")
	assert.NotEmpty(t, manager.Name())
	manager.open = func(string, string) (*sql.DB, error) { return nil, errors.New("open failed") }
	_, err = manager.Take(context.Background())
	assert.EqualError(t, err, "open failed")

	manager.open = func(string, string) (*sql.DB, error) { return db, nil }
	manager.ping = func(context.Context, *sql.DB) error { return errors.New("ping failed") }
	_, err = manager.Take(context.Background())
	assert.EqualError(t, err, "ping failed")

	manager.ping = func(context.Context, *sql.DB) error { return nil }
	manager.open = func(string, string) (*sql.DB, error) { return db2, nil }
	got, err := manager.Take(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, db2, got)
	again, err := manager.Take(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, got, again)
	assert.NoError(t, manager.Return(context.Background(), got))
	assert.NoError(t, manager.Shutdown(context.Background()))
	assert.NoError(t, (&OracleConnectionManager{}).Shutdown(context.Background()))

	fromConfig := NewOracleConnectionManagerFromConfig("h", 1, "svc", "u", "p")
	assert.Contains(t, fromConfig.ConnectionString, "oracle://")
}
