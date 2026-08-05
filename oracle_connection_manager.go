package oracle_storage

import (
	"context"
	"database/sql"
	"sync"

	go_ora "github.com/sijms/go-ora/v2"
	"github.com/storage-lock/go-storage"
)

// OracleConnectionManager Oracle 连接管理器
type OracleConnectionManager struct {
	ConnectionString string

	mu         sync.Mutex
	db         *sql.DB
	initialized bool
}

var _ storage.ConnectionManager[*sql.DB] = &OracleConnectionManager{}

// NewOracleConnectionManagerFromDSN 从数据源名称创建连接管理器
// DSN 格式: oracle://user:password@host:port/service_name
func NewOracleConnectionManagerFromDSN(dsn string) *OracleConnectionManager {
	return &OracleConnectionManager{
		ConnectionString: dsn,
	}
}

// NewOracleConnectionManagerFromConfig 从配置参数创建连接管理器
func NewOracleConnectionManagerFromConfig(host string, port int, service, username, password string) *OracleConnectionManager {
	// 构建 Oracle 连接字符串
	// 格式: oracle://username:password@host:port/service_name
	urlObj := go_ora.BuildUrl(host, port, service, username, password, nil)
	return &OracleConnectionManager{
		ConnectionString: urlObj,
	}
}

func (x *OracleConnectionManager) Name() string {
	return "oracle-connection-manager"
}

func (x *OracleConnectionManager) Take(ctx context.Context) (*sql.DB, error) {
	x.mu.Lock()
	defer x.mu.Unlock()

	if x.initialized {
		return x.db, nil
	}

	db, err := sql.Open("oracle", x.ConnectionString)
	if err != nil {
		return nil, err
	}

	// 配置连接池
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	// 测试连接
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

	x.db = db
	x.initialized = true
	return x.db, nil
}

func (x *OracleConnectionManager) Return(ctx context.Context, conn *sql.DB) error {
	return nil
}

func (x *OracleConnectionManager) Shutdown(ctx context.Context) error {
	x.mu.Lock()
	defer x.mu.Unlock()

	if x.db != nil {
		return x.db.Close()
	}
	return nil
}