package database

// DBClient is the interface for database operations (PostgreSQL)
type DBClient interface {
	From(table string) QueryBuilder
	Insert(table string, data map[string]interface{}) ([]byte, error)
	Update(table, keyColumn, keyValue string, data map[string]interface{}) ([]byte, error)
	Upsert(table string, data []map[string]interface{}, conflictColumn string) ([]byte, error)
	Delete(table, column, value string) error
}

// QueryBuilder is the interface for building queries
type QueryBuilder interface {
	Select(columns string) QueryBuilder
	Eq(column string, value interface{}) QueryBuilder
	Ilike(column string, pattern string) QueryBuilder
	Limit(limit int) QueryBuilder
	Single() QueryBuilder
	Order(column string, desc bool) QueryBuilder
	In(column string, values []string) QueryBuilder
	Gte(column, value string) QueryBuilder
	Lte(column, value string) QueryBuilder
	IsNull(column string) QueryBuilder
	Offset(n int) QueryBuilder
	Execute(result interface{}) error
}
