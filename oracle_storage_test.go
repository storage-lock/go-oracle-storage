package oracle_storage

import (
	"context"
	"os"
	"testing"

	storage_test_helper "github.com/storage-lock/go-storage-test-helper"
	"github.com/stretchr/testify/assert"
)

func TestOracleStorage(t *testing.T) {
	envName := "STORAGE_LOCK_ORACLE_DSN"
	dsn := os.Getenv(envName)
	if dsn == "" {
		t.Skip("STORAGE_LOCK_ORACLE_DSN not set, skipping Oracle test")
	}

	ctx := context.Background()
	cm := NewOracleConnectionManagerFromDSN(dsn)
	s, err := NewOracleStorage(ctx, &OracleStorageOptions{
		ConnectionManager: cm,
		TableName:         storage_test_helper.TestTableName,
	})
	assert.Nil(t, err)
	storage_test_helper.TestStorage(t, s)
}