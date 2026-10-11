package oracle_storage

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/storage-lock/go-storage"
	storage_lock "github.com/storage-lock/go-storage-lock"
	"github.com/stretchr/testify/assert"
)

type takeErrorManager struct{}

func (takeErrorManager) Name() string                          { return "take-error" }
func (takeErrorManager) Take(context.Context) (*sql.DB, error) { return nil, errors.New("take failed") }
func (takeErrorManager) Return(context.Context, *sql.DB) error { return nil }
func (takeErrorManager) Shutdown(context.Context) error        { return nil }

func TestOracleSqlProvider(t *testing.T) {
	provider := NewOracleSqlProvider()
	createSQL, args := provider.CreateTableSql(context.Background(), "locks")
	assert.Contains(t, createSQL, "lock_information_json_string")
	assert.Nil(t, args)
	nowSQL, args := provider.NowTimestampSql(context.Background(), "locks")
	assert.Contains(t, nowSQL, "SYSDATE")
	assert.Nil(t, args)
	listSQL, args := provider.ListLockInformationJsonStringSql(context.Background(), "locks")
	assert.Contains(t, listSQL, "locks")
	assert.Nil(t, args)
}

func TestOracleStorageOptions(t *testing.T) {
	opts := NewOracleStorageOptions().SetTableName("locks").SetConnectionManager(takeErrorManager{})
	assert.NoError(t, opts.Check())
	assert.EqualError(t, NewOracleStorageOptions().Check(), "ConnectionManager cannot be nil")
	assert.EqualError(t, NewOracleStorageOptions().SetConnectionManager(takeErrorManager{}).SetTableName("").Check(), "TableName cannot be empty")
	assert.Panics(t, func() { _, _ = NewOracleStorage(context.Background(), nil) })
	_, err := NewOracleStorage(context.Background(), NewOracleStorageOptions())
	assert.EqualError(t, err, "ConnectionManager cannot be nil")
}

func TestNewOracleStorage_ErrorsAndSuccess(t *testing.T) {
	_, err := NewOracleStorage(context.Background(), NewOracleStorageOptions().SetConnectionManager(takeErrorManager{}))
	assert.EqualError(t, err, "take failed")

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	assert.NoError(t, err)
	defer db.Close()
	mock.ExpectExec(".*").WillReturnError(errors.New("base failed"))
	_, err = NewOracleStorage(context.Background(), NewOracleStorageOptions().SetConnectionManager(storage.NewFixedSqlDBConnectionManager(db)))
	assert.EqualError(t, err, "base failed")

	db2, mock2, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	assert.NoError(t, err)
	defer db2.Close()
	mock2.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(0, 0))
	s, err := NewOracleStorage(context.Background(), NewOracleStorageOptions().SetConnectionManager(storage.NewFixedSqlDBConnectionManager(db2)))
	assert.NoError(t, err)
	assert.Equal(t, StorageName, s.GetName())
	assert.Contains(t, s.Capabilities(), storage.CapabilityWatch)
	_, err = s.Watch(context.Background(), "")
	assert.Error(t, err)

	info := &storage.LockInformation{OwnerId: "owner"}
	mock2.ExpectExec(".*").WillReturnError(errors.New("ORA-00001: unique constraint violated"))
	assert.ErrorIs(t, s.CreateWithVersion(context.Background(), "lock", 1, info), storage_lock.ErrLockAlreadyExists)
	mock2.ExpectExec(".*").WillReturnError(errors.New("other"))
	assert.EqualError(t, s.CreateWithVersion(context.Background(), "lock", 1, info), "other")
	mock2.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(1, 1))
	assert.NoError(t, s.CreateWithVersion(context.Background(), "lock", 1, info))
}
