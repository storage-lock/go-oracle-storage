package oracle_storage

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	sql_based_storage "github.com/storage-lock/go-sql-based-storage"
	"github.com/storage-lock/go-storage"
	storage_lock "github.com/storage-lock/go-storage-lock"
)

// OracleStorage 基于 Oracle 数据库的存储引擎实现
// Oracle 通过主键唯一约束和 WHERE 条件更新支持 CAS，
// 通过 SYSDATE 提供数据库服务器时间
type OracleStorage struct {
	*sql_based_storage.SqlBasedStorage
	options *OracleStorageOptions
}

var _ storage.Storage = &OracleStorage{}

func NewOracleStorage(ctx context.Context, options *OracleStorageOptions) (*OracleStorage, error) {
	if err := options.Check(); err != nil {
		return nil, err
	}

	baseStorageOption := sql_based_storage.NewSqlBasedStorageOptions().
		SetConnectionManager(options.ConnectionManager).
		SetSqlProvider(NewOracleSqlProvider()).
		SetTableFullName(options.TableName)
	baseStorage, err := sql_based_storage.NewSqlBasedStorage(baseStorageOption)
	if err != nil {
		return nil, err
	}

	s := &OracleStorage{
		SqlBasedStorage: baseStorage,
		options:         options,
	}

	if err := s.Init(ctx); err != nil {
		return nil, err
	}

	return s, nil
}

const StorageName = "oracle-storage"

func (x *OracleStorage) GetName() string {
	return StorageName
}

// Capabilities 声明 Oracle 存储支持的能力
// Oracle 通过主键唯一约束和 WHERE 条件更新天然支持 CAS，
// 通过 SYSDATE 提供数据库服务器时间
func (x *OracleStorage) Capabilities() []storage.StorageCapability {
	return []storage.StorageCapability{
		storage.CapabilityCAS,
		storage.CapabilityReliableTime,
	}
}

// CreateWithVersion Oracle 唯一约束冲突时返回 ErrLockAlreadyExists
func (x *OracleStorage) CreateWithVersion(ctx context.Context, lockId string, version storage.Version, lockInformation *storage.LockInformation) error {
	err := x.SqlBasedStorage.CreateWithVersion(ctx, lockId, version, lockInformation)
	if err != nil {
		// Oracle 唯一约束冲突错误码: ORA-00001
		// 错误格式: ORA-00001: unique constraint (SCHEMA.PK_TABLE) violated
		if strings.Contains(err.Error(), "ORA-00001") {
			return storage_lock.ErrLockAlreadyExists
		}
	}
	return err
}

// OracleStorageOptions Oracle 存储选项
type OracleStorageOptions struct {
	ConnectionManager storage.ConnectionManager[*sql.DB]
	TableName         string
}

func NewOracleStorageOptions() *OracleStorageOptions {
	return &OracleStorageOptions{
		TableName: storage.DefaultStorageTableName,
	}
}

func (x *OracleStorageOptions) SetConnectionManager(cm storage.ConnectionManager[*sql.DB]) *OracleStorageOptions {
	x.ConnectionManager = cm
	return x
}

func (x *OracleStorageOptions) SetTableName(name string) *OracleStorageOptions {
	x.TableName = name
	return x
}

func (x *OracleStorageOptions) Check() error {
	if x.ConnectionManager == nil {
		return errors.New("ConnectionManager cannot be nil")
	}
	if x.TableName == "" {
		return errors.New("TableName cannot be empty")
	}
	return nil
}
