package aaronsql

import (
	"fmt"
	"reflect"
	"time"
)

type PostgresDataBase struct {
	DataBase
}

type PostgresColumn struct {
	BaseColumn
}

// GetName returns the name of the database type.
func (postgres *PostgresDataBase) Name() DBName {
	return PostgresDB
}

// GetDB returns the underlying sql.DB instance.
func (postgres *PostgresDataBase) GetDB() *DataBase {
	return &postgres.DataBase
}

// GetTables returns the DDL information for all tables in the database.
func (postgres *PostgresDataBase) GetTables() ([]Table, error) {
	ret := make([]Table, 0)
	tableNames, err := postgres.getTableNames()
	if err != nil {
		return nil, fmt.Errorf("failed to get table names: %w", err)
	}
	for _, tableName := range tableNames {
		columns, err := postgres.getColumnInfo(tableName)
		if err != nil {
			return nil, fmt.Errorf("failed to get columns for table %s: %w", tableName, err)
		}
		table := Table{
			name:         tableName,
			columns:      columns,
			db:           postgres,
			extraOptions: make(map[string]string),
		}
		ret = append(ret, table)
	}
	return ret, nil
}

func (postgres *PostgresDataBase) GetCreateTableSQL(tableName string, columns []ColumnInterface) string {
	sql := fmt.Sprintf("CREATE TABLE %s (", tableName)
	for i, col := range columns {
		sql += fmt.Sprintf("%s %s", col.Name(), col.Type())
		if i < len(columns)-1 {
			sql += ", "
		}
	}
	sql += ");"
	return sql
}

func (postgres *PostgresDataBase) IsSupportForeignKeys() bool {
	return true
}

func (postgres *PostgresDataBase) GetTablesColumns(t TableInterface) ([]ColumnInterface, error) {
	ret := make([]ColumnInterface, 0)
	for _, col := range t.Columns() {
		if col != nil {
			ret = append(ret, col)
		}
	}
	if len(ret) == 0 {
		return nil, fmt.Errorf("no columns found for table %s", t.Name())
	}
	return ret, nil
}

func (postgres *PostgresDataBase) GetColumnDefinitionByType(fieldType reflect.Type, columnName string, tag map[string]string, isPointer bool) (ColumnInterface, error) {
	retCol := PostgresColumn{}
	retCol.name = columnName
	if isPointer {
		retCol.isPointer = true
	} else {
		retCol.isPointer = false
	}

	// Handle pointer types by getting the underlying type
	actualType := fieldType
	if fieldType.Kind() == reflect.Ptr {
		actualType = fieldType.Elem()
		retCol.isPointer = true
	}

	switch actualType.Kind() {
	case reflect.String:
		retCol.sqlType = "TEXT"
	case reflect.Int8:
		retCol.sqlType = "SMALLINT"
	case reflect.Int16:
		retCol.sqlType = "SMALLINT"
	case reflect.Int32, reflect.Int:
		retCol.sqlType = "INTEGER"
	case reflect.Int64:
		retCol.sqlType = "BIGINT"
	case reflect.Uint8:
		retCol.sqlType = "SMALLINT"
	case reflect.Uint16:
		retCol.sqlType = "INTEGER"
	case reflect.Uint32, reflect.Uint:
		retCol.sqlType = "BIGINT"
	case reflect.Uint64:
		retCol.sqlType = "BIGINT" // Note: PostgreSQL doesn't have unsigned types
	case reflect.Float32:
		retCol.sqlType = "REAL"
	case reflect.Float64:
		retCol.sqlType = "DOUBLE PRECISION"
	case reflect.Bool:
		retCol.sqlType = "BOOLEAN"
	case reflect.Slice:
		if actualType == reflect.TypeOf([]byte{}) {
			retCol.sqlType = "BYTEA"
		} else {
			return nil, fmt.Errorf("unsupported slice type: %s", actualType.String())
		}
	case reflect.Struct:
		if actualType == reflect.TypeOf(time.Time{}) {
			retCol.sqlType = "TIMESTAMP WITH TIME ZONE"
		} else {
			return nil, fmt.Errorf("unsupported struct type: %s", actualType.Name())
		}
	default:
		return nil, fmt.Errorf("unsupported field type: %s", actualType.Kind().String())
	}
	if defaultValue, ok := tag["default"]; ok {
		retCol.defaultString = defaultValue
	} else {
		retCol.defaultString = ""
	}
	if nullable, ok := tag["nullable"]; ok && nullable == "true" {
		retCol.isNullable = true
	} else {
		retCol.isNullable = false
	}
	if primaryKey, ok := tag["primary_key"]; ok && primaryKey == "true" {
		retCol.isPrimaryKey = true
	} else {
		retCol.isPrimaryKey = false
	}
	if unique, ok := tag["unique"]; ok && unique == "true" {
		retCol.isUnique = true
	} else {
		retCol.isUnique = false
	}
	if index, ok := tag["index"]; ok && index == "true" {
		retCol.isIndex = true
	} else {
		retCol.isIndex = false
	}
	if allowZero, ok := tag["allow_zero"]; ok && allowZero == "true" {
		retCol.isAllowZero = true
	} else {
		retCol.isAllowZero = false
	}
	retCol.tags = tag
	retCol.columnIndex = -1 // Default value, can be set later if needed
	if name, ok := tag["name"]; ok {
		retCol.name = name
	} else {
		retCol.name = columnName
	}
	return &retCol, nil
}

func (postgres *PostgresDataBase) DropTableSql(tableName string) string {
	return fmt.Sprintf("DROP TABLE IF EXISTS %s;", tableName)
}

func (postgres *PostgresDataBase) CanInsert() bool {
	return true
}

func (postgres *PostgresDataBase) CanInsertOrUpdate() bool {
	return true
}

func (postgres *PostgresDataBase) CanUpdate() bool {
	return true
}

func (postgres *PostgresDataBase) CanReturnRowsAffected() bool {
	return true
}

func (postgres *PostgresDataBase) InsertSqlTemplate() string {
	tpl := ("INSERT INTO {{.TableName}} ({{.Columns}}) VALUES ({{.Values}});")
	return tpl
}

func (postgres *PostgresDataBase) UpdateSqlTemplate() string {
	tpl := ("UPDATE {{.TableName}} SET {{.Updates}} WHERE {{.Conditions}};")
	return tpl
}

func (postgres *PostgresDataBase) InsertOrUpdateSqlTemplate() string {
	tpl := ("INSERT INTO {{.TableName}} ({{.Columns}}) VALUES ({{.Values}}) ON CONFLICT ({{.ConflictColumns}}) DO UPDATE SET {{.Updates}};")
	return tpl
}

func (postgres *PostgresDataBase) CanRenameTable() bool {
	return true
}

func (postgres *PostgresDataBase) GetTableDDL(tableName string) (*Table, error) {
	table := &Table{
		name:        tableName,
		columns:     make([]ColumnInterface, 0),
		indexes:     make([]TableIndex, 0),
		constraints: make([]TableForeignKey, 0),
		db:          postgres,
	}
	columnMap := make(map[string]*PostgresColumn)
	columnQuery := `
		SELECT
			column_name,
			udt_name,
			is_nullable,
			column_default
		FROM
			information_schema.columns
		WHERE
			table_schema = $1 AND table_name = $2
		ORDER BY
			ordinal_position;
	`
	rows, err := postgres.db.Query(columnQuery, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var colName, dataType, isNullable, defaultValue string
		if err := rows.Scan(&colName, &dataType, &isNullable, &defaultValue); err != nil {
			return nil, err
		}

		column := &PostgresColumn{
			BaseColumn: BaseColumn{
				name:          colName,
				sqlType:       dataType,
				isNullable:    isNullable == "YES",
				defaultString: defaultValue,
			},
		}
		columnMap[colName] = column
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(columnMap) == 0 {
		return nil, fmt.Errorf("no columns found for table %s", tableName)
	}

	table.columns = make([]ColumnInterface, 0, len(columnMap))
	for _, col := range columnMap {
		table.columns = append(table.columns, col)
	}
	return table, nil
}

func (postgres *PostgresDataBase) getTableNames() ([]string, error) {
	query := `
		SELECT tablename
		FROM pg_catalog.pg_tables
		WHERE schemaname != 'pg_catalog' AND schemaname != 'information_schema';
	`
	rows, err := postgres.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}
		tables = append(tables, tableName)
	}
	return tables, nil
}

func (postgres *PostgresDataBase) getColumnInfo(tableName string) ([]ColumnInterface, error) {
	query := `
		SELECT column_name, data_type, is_nullable, column_default
		FROM information_schema.columns
		WHERE table_name = $1
		ORDER BY ordinal_position;
	`
	rows, err := postgres.db.Query(query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []ColumnInterface
	for rows.Next() {
		var colName, dataType, isNullable, defaultValue string
		if err := rows.Scan(&colName, &dataType, &isNullable, &defaultValue); err != nil {
			return nil, err
		}

		column := &PostgresColumn{
			BaseColumn: BaseColumn{
				name:          colName,
				sqlType:       dataType,
				isNullable:    isNullable == "YES",
				defaultString: defaultValue,
			},
		}
		columns = append(columns, column)

	}
	return columns, nil
}

func (postgres *PostgresDataBase) CreateIndexSqlTemplate() string {
	return "CREATE INDEX IF NOT EXISTS {{.IndexName}} ON {{.TableName}} ({{.Columns}});"
}

func (postgres *PostgresDataBase) DropIndexSqlTemplate() string {
	return "DROP INDEX IF EXISTS {{.IndexName}};"
}

func (postgres *PostgresDataBase) CreateColumnSqlTemplate() string {
	return "ALTER TABLE {{.TableName}} ADD COLUMN {{.ColumnName}} {{.ColumnType}};"
}

func (postgres *PostgresDataBase) UpdateColumnSqlTemplate() string {
	return "ALTER TABLE {{.TableName}} ALTER COLUMN {{.ColumnName}} SET DATA TYPE {{.ColumnType}};"
}

