package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
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

// relationInfo stores information about a Supabase-style relationship
type relationInfo struct {
	alias      string   // e.g., "assignee"
	table      string   // e.g., "employees"
	foreignKey string   // e.g., "tasks_assignee_id_fkey" (optional)
	columns    []string // e.g., ["id", "name"]
}

// PostgresQueryBuilder helps build SQL queries
type PostgresQueryBuilder struct {
	client       *PostgresClient
	table        string
	columns      string
	relations    []relationInfo
	whereClauses []string
	whereArgs    []interface{}
	limitVal     int
	offsetVal    int
	orderBy      string
	single       bool
}

// From starts a query on a table
func (c *PostgresClient) From(table string) QueryBuilder {
	return &PostgresQueryBuilder{
		client:  c,
		table:   table,
		columns: "*",
	}
}

// parseSupabaseColumns parses Supabase-style column syntax and extracts relations
// Example: "*, assignee:employees!tasks_assignee_id_fkey(id, name), project:projects(id, name)"
// Also supports short form: "*, employees(name, position)" where table name is used as alias
func parseSupabaseColumns(columns string) (string, []relationInfo) {
	var relations []relationInfo
	var cleanColumns []string

	// Regex to match relationship syntax with alias: alias:table!foreignkey(cols) or alias:table(cols)
	// Pattern: word:word(!word)?(...)
	aliasRelationRegex := regexp.MustCompile(`(\w+):(\w+)(?:!(\w+))?\(([^)]+)\)`)

	// Regex to match short form: table(cols) - no alias prefix
	// Pattern: word(...) but NOT preceded by :
	shortRelationRegex := regexp.MustCompile(`(?:^|[,\s])(\w+)\(([^)]+)\)`)

	// Find all relations with alias
	aliasMatches := aliasRelationRegex.FindAllStringSubmatch(columns, -1)
	for _, match := range aliasMatches {
		cols := strings.Split(match[4], ",")
		for i := range cols {
			cols[i] = strings.TrimSpace(cols[i])
		}
		relations = append(relations, relationInfo{
			alias:      match[1],
			table:      match[2],
			foreignKey: match[3],
			columns:    cols,
		})
	}

	// Find short form relations (table as alias)
	shortMatches := shortRelationRegex.FindAllStringSubmatch(columns, -1)
	for _, match := range shortMatches {
		tableName := match[1]
		// Skip if this table was already matched with alias syntax
		alreadyMatched := false
		for _, rel := range relations {
			if rel.table == tableName {
				alreadyMatched = true
				break
			}
		}
		if alreadyMatched {
			continue
		}

		cols := strings.Split(match[2], ",")
		for i := range cols {
			cols[i] = strings.TrimSpace(cols[i])
		}
		relations = append(relations, relationInfo{
			alias:      tableName, // Use table name as alias
			table:      tableName,
			foreignKey: "",
			columns:    cols,
		})
	}

	// Remove relation syntax from columns string
	cleanStr := aliasRelationRegex.ReplaceAllString(columns, "")
	cleanStr = shortRelationRegex.ReplaceAllString(cleanStr, " ")

	// Split by comma and clean up
	parts := strings.Split(cleanStr, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			cleanColumns = append(cleanColumns, part)
		}
	}

	// If nothing left, default to *
	if len(cleanColumns) == 0 {
		cleanColumns = append(cleanColumns, "*")
	}

	return strings.Join(cleanColumns, ", "), relations
}

// Select specifies which columns to select
func (qb *PostgresQueryBuilder) Select(columns string) QueryBuilder {
	cleanCols, relations := parseSupabaseColumns(columns)
	qb.columns = cleanCols
	qb.relations = relations
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
	qb.single = true
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

	// Fetch related data for each relation
	if len(qb.relations) > 0 {
		results = qb.fetchRelations(results)
	}

	// Marshal and unmarshal to convert to target type
	// If single mode and results exist, return first element
	if qb.single {
		if len(results) == 0 {
			return fmt.Errorf("no rows found")
		}
		jsonData, err := json.Marshal(results[0])
		if err != nil {
			return err
		}
		return json.Unmarshal(jsonData, result)
	}

	jsonData, err := json.Marshal(results)
	if err != nil {
		return err
	}

	return json.Unmarshal(jsonData, result)
}

// fetchRelations fetches related data for each result row
func (qb *PostgresQueryBuilder) fetchRelations(results []map[string]interface{}) []map[string]interface{} {
	for _, rel := range qb.relations {
		// Determine the foreign key column name
		fkColumn := singularize(rel.alias) + "_id"
		if rel.foreignKey != "" {
			// Parse foreign key name like "tasks_assignee_id_fkey" to get "assignee_id"
			parts := strings.Split(rel.foreignKey, "_")
			if len(parts) >= 3 {
				// Try to extract column name (e.g., "assignee_id" from "tasks_assignee_id_fkey")
				fkParts := []string{}
				for i := 1; i < len(parts)-1; i++ { // Skip table prefix and "fkey" suffix
					fkParts = append(fkParts, parts[i])
				}
				if len(fkParts) > 0 {
					fkColumn = strings.Join(fkParts, "_")
				}
			}
		}

		// Collect all foreign key values
		fkValues := make(map[string]bool)
		for _, row := range results {
			if fkVal, ok := row[fkColumn]; ok && fkVal != nil {
				fkValues[fmt.Sprintf("%v", fkVal)] = true
			}
		}

		if len(fkValues) == 0 {
			// No foreign keys, set all relations to null
			for i := range results {
				results[i][rel.alias] = nil
			}
			continue
		}

		// Fetch related data
		relatedData := qb.fetchRelatedData(rel, fkValues)

		// Merge related data into results
		for i, row := range results {
			if fkVal, ok := row[fkColumn]; ok && fkVal != nil {
				fkStr := fmt.Sprintf("%v", fkVal)
				if related, found := relatedData[fkStr]; found {
					results[i][rel.alias] = related
				} else {
					results[i][rel.alias] = nil
				}
			} else {
				results[i][rel.alias] = nil
			}
		}
	}

	return results
}

// fetchRelatedData fetches data from a related table
func (qb *PostgresQueryBuilder) fetchRelatedData(rel relationInfo, fkValues map[string]bool) map[string]map[string]interface{} {
	relatedData := make(map[string]map[string]interface{})

	if len(fkValues) == 0 {
		return relatedData
	}

	// Build columns list
	cols := strings.Join(rel.columns, ", ")
	if !containsColumn(rel.columns, "id") {
		cols = "id, " + cols
	}

	// Build IN clause
	fkList := make([]string, 0, len(fkValues))
	for fk := range fkValues {
		fkList = append(fkList, fk)
	}

	// Build placeholders
	placeholders := make([]string, len(fkList))
	args := make([]interface{}, len(fkList))
	for i, fk := range fkList {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = fk
	}

	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE id IN (%s)",
		cols,
		rel.table,
		strings.Join(placeholders, ", "),
	)

	rows, err := qb.client.db.Query(query, args...)
	if err != nil {
		return relatedData
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return relatedData
	}

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}

		row := make(map[string]interface{})
		var id string
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
				if col == "id" {
					id = string(b)
				}
			} else {
				row[col] = val
				if col == "id" {
					id = fmt.Sprintf("%v", val)
				}
			}
		}
		if id != "" {
			relatedData[id] = row
		}
	}

	return relatedData
}

// containsColumn checks if a column list contains a specific column
func containsColumn(cols []string, col string) bool {
	for _, c := range cols {
		if c == col {
			return true
		}
	}
	return false
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

	rows, err := c.db.Query(query, values...)
	if err != nil {
		return nil, fmt.Errorf("insert failed: %w", err)
	}
	defer rows.Close()

	// Get actual column names from result
	columnNames, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	for rows.Next() {
		scanDest := make([]interface{}, len(columnNames))
		for i := range scanDest {
			var v interface{}
			scanDest[i] = &v
		}

		if err := rows.Scan(scanDest...); err != nil {
			return nil, err
		}

		result := make(map[string]interface{})
		for i, col := range columnNames {
			val := *(scanDest[i].(*interface{}))
			// Convert []byte to string for JSON compatibility
			if b, ok := val.([]byte); ok {
				result[col] = string(b)
			} else {
				result[col] = val
			}
		}
		results = append(results, result)
	}

	return json.Marshal(results)
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
		"UPDATE %s SET %s WHERE %s = $%d RETURNING *",
		table,
		strings.Join(setClauses, ", "),
		keyColumn,
		i,
	)

	rows, err := c.db.Query(query, values...)
	if err != nil {
		return nil, fmt.Errorf("update failed: %w", err)
	}
	defer rows.Close()

	// Get actual column names from result
	columnNames, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	for rows.Next() {
		scanDest := make([]interface{}, len(columnNames))
		for i := range scanDest {
			var v interface{}
			scanDest[i] = &v
		}

		if err := rows.Scan(scanDest...); err != nil {
			return nil, err
		}

		result := make(map[string]interface{})
		for i, col := range columnNames {
			val := *(scanDest[i].(*interface{}))
			if b, ok := val.([]byte); ok {
				result[col] = string(b)
			} else {
				result[col] = val
			}
		}
		results = append(results, result)
	}

	return json.Marshal(results)
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

// singularize converts plural table names to singular for foreign key lookups
func singularize(name string) string {
	// Special cases for common table names
	specialCases := map[string]string{
		"employees":          "employee",
		"projects":           "project",
		"meetings":           "meeting",
		"meeting_categories": "category",
		"tasks":              "task",
		"categories":         "category",
		"conversations":      "conversation",
		"messages":           "message",
		"tags":               "tag",
		"files":              "file",
	}
	if singular, ok := specialCases[name]; ok {
		return singular
	}
	// Simple pluralization rules
	if strings.HasSuffix(name, "ies") {
		return name[:len(name)-3] + "y"
	}
	if strings.HasSuffix(name, "s") && !strings.HasSuffix(name, "ss") {
		return name[:len(name)-1]
	}
	return name
}
