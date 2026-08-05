package oracle_storage

import (
	"context"
	sql_based_storage "github.com/storage-lock/go-sql-based-storage"
)

// OracleSqlProvider Oracle 数据库的 SQL 方言提供者
// Oracle 使用 PL/SQL 方言，与标准 SQL 有一些差异：
//   - 时间函数：SYSDATE 或 CURRENT_TIMESTAMP（而非 NOW()）
//   - 字符串连接：使用 || 而非 CONCAT()
//   - 限制行数：使用 ROWNUM 而非 LIMIT
type OracleSqlProvider struct {
	*sql_based_storage.Sql92Provider
}

func NewOracleSqlProvider() *OracleSqlProvider {
	return &OracleSqlProvider{
		Sql92Provider: sql_based_storage.NewSql92Provider(),
	}
}

// CreateTableSql Oracle 创建表的 SQL
// Oracle 不支持 AUTO_INCREMENT，使用序列和触发器或直接插入值
func (x *OracleSqlProvider) CreateTableSql(ctx context.Context, tableFullName string) (string, []any) {
	// Oracle 表名通常用大写，但保留用户指定的格式
	sql := `CREATE TABLE ` + tableFullName + ` (
		lock_id VARCHAR2(255) PRIMARY KEY,
		version NUMBER(20) NOT NULL,
		json_data CLOB NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`
	return sql, nil
}

// NowTimestampSql Oracle 获取当前时间戳
// Oracle 使用 SYSDATE 或 CURRENT_TIMESTAMP，返回的是 TIMESTAMP 类型
// 需要转换为 Unix 时间戳
func (x *OracleSqlProvider) NowTimestampSql(ctx context.Context, tableFullName string) (string, []any) {
	// Oracle: 使用 (SYSDATE - TO_DATE('1970-01-01', 'YYYY-MM-DD')) * 86400 获取 Unix 时间戳
	// 或者使用更简单的方式：抽取 SYSDATE 的 Unix 时间戳
	// 注意：Oracle 的日期运算单位是天，所以需要乘以 86400（秒/天）
	return `SELECT CAST((SYSDATE - TO_DATE('1970-01-01', 'YYYY-MM-DD')) * 86400 AS NUMBER(20)) FROM DUAL`, nil
}

// ListLockInformationJsonStringSql Oracle 列出所有锁
// Oracle 不支持 LIMIT，使用 ROWNUM 或者 FETCH FIRST
func (x *OracleSqlProvider) ListLockInformationJsonStringSql(ctx context.Context, tableFullName string) (string, []any) {
	return `SELECT json_data FROM ` + tableFullName, nil
}