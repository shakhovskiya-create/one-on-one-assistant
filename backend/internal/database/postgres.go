package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// PostgresClient wraps database/sql for PostgreSQL with Supabase-compatible API
type PostgresClient struct {
	db *sql.DB
}

// NewPostgresClient creates a new PostgreSQL client
func NewPostgresClient(connectionString string) (*PostgresClient, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &PostgresClient{db: db}, nil
}

// PostgresQueryBuilder helps build SQL queries
type PostgresQueryBuilder struct {
	client       *PostgresClient
	table        string
	columns      string
	whereClauses []string
	whereArgs    []interface{}
	limitVal     int
	offsetVal    int
	orderBy      string
}

// From starts a query on a table
func (c *PostgresClient) From(table string) QueryBuilder {
	return &PostgresQueryBuilder{
		client:  c,
		table:   table,
		columns: "*",
	}
}

// Select specifies which columns to select
func (qb *PostgresQueryBuilder) Select(columns string) QueryBuilder {
	qb.columns = columns
	return qb
}

// Eq adds an equality WHERE clause
func (qb *PostgresQueryBuilder) Eq(column string, value interface{}) QueryBuilder {
	qb.whereClauses = append(qb.whereClauses, fmt.Sprintf("%s = $%d", column, len(qb.whereArgs)+1))
	qb.whereArgs = append(qb.whereArgs, value)
	return qb
}

// Ilike adds a case-insensitive LIKE clause
func (qb *PostgresQueryBuilder) Ilike(column string, pattern string) QueryBuilder {
	qb.whereClauses = append(qb.whereClauses, fmt.Sprintf("%s ILIKE $%d", column, len(qb.whereArgs)+1))
	qb.whereArgs = append(qb.whereArgs, pattern)
	return qb
}

// Order adds ORDER BY clause
func (qb *PostgresQueryBuilder) Order(column string, desc bool) QueryBuilder {
	order := "ASC"
	if desc {
		order = "DESC"
	}
	qb.orderBy = fmt.Sprintf("%s %s", column, order)
	return qb
}

// In adds IN clause
func (qb *PostgresQueryBuilder) In(column string, values []string) QueryBuilder {
	placeholders := make([]string, len(values))
	for i, v := range values {
		placeholders[i] = fmt.Sprintf("$%d", len(qb.whereArgs)+1)
		qb.whereArgs = append(qb.whereArgs, v)
	}
	qb.whereClauses = append(qb.whereClauses, fmt.Sprintf("%s IN (%s)", column, strings.Join(placeholders, ", ")))
	return qb
}

// Gte adds >= clause
func (qb *PostgresQueryBuilder) Gte(column, value string) QueryBuilder {
	qb.whereClauses = append(qb.whereClauses, fmt.Sprintf("%s >= $%d", column, len(qb.whereArgs)+1))
	qb.whereArgs = append(qb.whereArgs, value)
	return qb
}

// Lte adds <= clause
func (qb *PostgresQueryBuilder) Lte(column, value string) QueryBuilder {
	qb.whereClauses = append(qb.whereClauses, fmt.Sprintf("%s <= $%d", column, len(qb.whereArgs)+1))
	qb.whereArgs = append(qb.whereArgs, value)
	return qb
}

// IsNull adds IS NULL clause
func (qb *PostgresQueryBuilder) IsNull(column string) QueryBuilder {
	qb.whereClauses = append(qb.whereClauses, fmt.Sprintf("%s IS NULL", column))
	return qb
}

// Offset skips results
func (qb *PostgresQueryBuilder) Offset(n int) QueryBuilder {
	qb.offsetVal = n
	return qb
}

// Limit sets the LIMIT
func (qb *PostgresQueryBuilder) Limit(limit int) QueryBuilder {
	qb.limitVal = limit
	return qb
}

// Single executes query expecting single result
func (qb *PostgresQueryBuilder) Single() QueryBuilder {
	qb.limitVal = 1
	return qb
}

// Execute executes the query and scans results into the provided slice
func (qb *PostgresQueryBuilder) Execute(result interface{}) error {
	query := fmt.Sprintf("SELECT %s FROM %s", qb.columns, qb.table)

	if len(qb.whereClauses) > 0 {
		query += " WHERE " + strings.Join(qb.whereClauses, " AND ")
	}

	if qb.orderBy != "" {
		query += " ORDER BY " + qb.orderBy
	}

	if qb.limitVal > 0 {
		query += fmt.Sprintf(" LIMIT %d", qb.limitVal)
	}

	if qb.offsetVal > 0 {
		query += fmt.Sprintf(" OFFSET %d", qb.offsetVal)
	}

	rows, err := qb.client.db.Query(query, qb.whereArgs...)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	// Parse results into slice of maps
	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return err
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			// Convert []byte to string for JSON compatibility
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		results = append(results, row)
	}

	// Marshal and unmarshal to convert to target type
	jsonData, err := json.Marshal(results)
	if err != nil {
		return err
	}

	return json.Unmarshal(jsonData, result)
}

// Insert inserts data into table
func (c *PostgresClient) Insert(table string, data map[string]interface{}) ([]byte, error) {
	columns := []string{}
	placeholders := []string{}
	values := []interface{}{}

	i := 1
	for col, val := range data {
		columns = append(columns, col)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		values = append(values, val)
		i++
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s) RETURNING *",
		table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	row := c.db.QueryRow(query, values...)

	// Scan result
	columnNames := []string{}
	for col := range data {
		columnNames = append(columnNames, col)
	}
	columnNames = append(columnNames, "id") // Assuming id is returned

	result := make(map[string]interface{})
	scanDest := make([]interface{}, len(columnNames))
	for i := range scanDest {
		var v interface{}
		scanDest[i] = &v
	}

	if err := row.Scan(scanDest...); err != nil {
		return nil, err
	}

	for i, col := range columnNames {
		result[col] = scanDest[i]
	}

	return json.Marshal([]map[string]interface{}{result})
}

// Update updates data in table
func (c *PostgresClient) Update(table, keyColumn, keyValue string, data map[string]interface{}) ([]byte, error) {
	setClauses := []string{}
	values := []interface{}{}

	i := 1
	for col, val := range data {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", col, i))
		values = append(values, val)
		i++
	}
	values = append(values, keyValue)

	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE %s = $%d",
		table,
		strings.Join(setClauses, ", "),
		keyColumn,
		i,
	)

	_, err := c.db.Exec(query, values...)
	return nil, err
}

// Upsert upserts data (not implemented yet, just inserts)
func (c *PostgresClient) Upsert(table string, data []map[string]interface{}, conflictColumn string) ([]byte, error) {
	ctx := context.Background()
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	for _, row := range data {
		columns := []string{}
		placeholders := []string{}
		values := []interface{}{}
		updateClauses := []string{}

		i := 1
		for col, val := range row {
			columns = append(columns, col)
			placeholders = append(placeholders, fmt.Sprintf("$%d", i))
			values = append(values, val)
			if col != conflictColumn {
				updateClauses = append(updateClauses, fmt.Sprintf("%s = EXCLUDED.%s", col, col))
			}
			i++
		}

		query := fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES (%s) ON CONFLICT (%s) DO UPDATE SET %s",
			table,
			strings.Join(columns, ", "),
			strings.Join(placeholders, ", "),
			conflictColumn,
			strings.Join(updateClauses, ", "),
		)

		if _, err := tx.Exec(query, values...); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return nil, nil
}

// Delete deletes data from a table
func (c *PostgresClient) Delete(table, column, value string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE %s = $1", table, column)
	_, err := c.db.Exec(query, value)
	return err
}

// StorageUpload - not implemented for PostgreSQL (файлы хранятся в файловой системе или S3)
func (c *PostgresClient) StorageUpload(bucket, path string, data io.Reader, contentType string) (string, error) {
	return "", fmt.Errorf("storage operations not supported with direct PostgreSQL connection")
}

// StorageDownload - not implemented for PostgreSQL
func (c *PostgresClient) StorageDownload(bucket, path string) ([]byte, string, error) {
	return nil, "", fmt.Errorf("storage operations not supported with direct PostgreSQL connection")
}

// StorageDelete - not implemented for PostgreSQL
func (c *PostgresClient) StorageDelete(bucket, path string) error {
	return fmt.Errorf("storage operations not supported with direct PostgreSQL connection")
}

// StorageGetPublicURL - not implemented for PostgreSQL
func (c *PostgresClient) StorageGetPublicURL(bucket, path string) string {
	return ""
}

// Close closes the database connection
func (c *PostgresClient) Close() error {
	return c.db.Close()
}
