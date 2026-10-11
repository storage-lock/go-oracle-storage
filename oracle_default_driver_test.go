package oracle_storage

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type defaultDriver struct{}

func (defaultDriver) Open(string) (driver.Conn, error) { return defaultConn{}, nil }

type defaultConn struct{}

func (defaultConn) Prepare(string) (driver.Stmt, error) { return nil, errors.New("unsupported") }
func (defaultConn) Close() error                        { return nil }
func (defaultConn) Begin() (driver.Tx, error)           { return nil, errors.New("unsupported") }
func (defaultConn) Ping(context.Context) error          { return errors.New("default ping failed") }

func init() {
	oracleDriverName = "oracle-coverage"
	sql.Register(oracleDriverName, defaultDriver{})
}

func TestConnectionManagerDefaultDriver(t *testing.T) {
	manager := NewOracleConnectionManagerFromDSN("oracle://user:pass@127.0.0.1:1/svc")
	_, err := manager.Take(context.Background())
	assert.EqualError(t, err, "default ping failed")
}
