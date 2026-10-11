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
//
// 注意：列名必须与 Sql92Provider 未被本 Provider 覆盖的
// CreateWithVersionSql/UpdateWithVersionSql/DeleteWithVersionSql/FindLockInformationJsonStringByIdSql
// 保持一致（lock_id/owner_id/version/lock_information_json_string）——这几个方法本 Provider
// 没有重写，会直接继承基类实现，如果表结构的列名对不上，CAS 相关的每条 SQL 都会在真实 Oracle 上
// 报"invalid identifier"（ORA-00904）。修复前这里曾经是 lock_id/version/json_data/created_at，
// 缺 owner_id 且 JSON 列名对不上，导致本存储在真实 Oracle 上完全不可用（未接入真实驱动测试，
// 靠走读代码发现）。
func (x *OracleSqlProvider) CreateTableSql(ctx context.Context, tableFullName string) (string, []any) {
	// Oracle 表名通常用大写，但保留用户指定的格式
	sql := `CREATE TABLE ` + tableFullName + ` (
		lock_id VARCHAR2(255) PRIMARY KEY,
		owner_id VARCHAR2(255) NOT NULL,
		version NUMBER(20) NOT NULL,
		lock_information_json_string CLOB NOT NULL
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
// 修复：原来查询的是不存在的 json_data 列（与 CreateTableSql 的列名不一致，见上），改为实际的
// lock_information_json_string 列。
func (x *OracleSqlProvider) ListLockInformationJsonStringSql(ctx context.Context, tableFullName string) (string, []any) {
	return `SELECT lock_information_json_string FROM ` + tableFullName, nil
}