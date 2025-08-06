package aaronsql

import (
	"fmt"
	"reflect"
	"time"
)

type MariaDBDataBase struct {
	DataBase
}

type MariaDBColumn struct {
	BaseColumn
}

// Name returns the name of the database type.
func (mariadb *MariaDBDataBase) Name() DBName {
	return MariaDB
}

// GetDB returns the underlying sql.DB instance.
func (mariadb *MariaDBDataBase) GetDB() *DataBase {
	return &mariadb.DataBase
}

// GetTables returns the DDL information for all tables in the database.
func (mariadb *MariaDBDataBase) GetTables() ([]Table, error) {
	ret := make([]Table, 0)
	tableNames, err := mariadb.getTableNames()
	if err != nil {
		return nil, fmt.Errorf("failed to get table names: %w", err)
	}
	for _, tableName := range tableNames {
		columns, err := mariadb.getColumnInfo(tableName)
		if err != nil {
			return nil, fmt.Errorf("failed to get columns for table %s: %w", tableName, err)
		}
		table := Table{
			name:         tableName,
			columns:      columns,
			db:           mariadb,
			extraOptions: make(map[string]string),
		}
		ret = append(ret, table)
	}
	return ret, nil
}

func (mariadb *MariaDBDataBase) GetCreateTableSQL(tableName string, columns []ColumnInterface) string {
	sql := fmt.Sprintf("CREATE TABLE `%s` (", tableName)
	primaryKeys := make([]string, 0)

	for i, col := range columns {
		sql += fmt.Sprintf("`%s` %s", col.Name(), col.Type())

		if !col.Nullable() {
			sql += " NOT NULL"
		}

		if col.Default() != "" {
			sql += fmt.Sprintf(" DEFAULT %s", col.Default())
		}

		if col.IsAutoIncrement() {
			sql += " AUTO_INCREMENT"
		}

		if col.IsPrimaryKey() {
			primaryKeys = append(primaryKeys, fmt.Sprintf("`%s`", col.Name()))
		}

		if i < len(columns)-1 {
			sql += ", "
		}
	}

	if len(primaryKeys) > 0 {
		sql += ", PRIMARY KEY (" + primaryKeys[0]
		for i := 1; i < len(primaryKeys); i++ {
			sql += ", " + primaryKeys[i]
		}
		sql += ")"
	}

	sql += ") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;"
	return sql
}

func (mariadb *MariaDBDataBase) IsSupportForeignKeys() bool {
	return true
}

func (mariadb *MariaDBDataBase) GetTablesColumns(t TableInterface) ([]ColumnInterface, error) {
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

func (mariadb *MariaDBDataBase) GetColumnDefinitionByType(fieldType reflect.Type, columnName string, tag map[string]string, isPointer bool) (ColumnInterface, error) {
	retCol := MariaDBColumn{}
	retCol.name = columnName
	retCol.isPointer = isPointer

	// Handle pointer types by getting the underlying type
	actualType := fieldType
	if fieldType.Kind() == reflect.Ptr {
		actualType = fieldType.Elem()
		retCol.isPointer = true
	}

	// Set default nullable based on pointer type
	retCol.isNullable = isPointer

	switch actualType.Kind() {
	case reflect.String:
		if length, ok := tag["length"]; ok {
			retCol.sqlType = fmt.Sprintf("VARCHAR(%s)", length)
		} else {
			retCol.sqlType = "TEXT"
		}
	case reflect.Int8:
		retCol.sqlType = "TINYINT"
	case reflect.Int16:
		retCol.sqlType = "SMALLINT"
	case reflect.Int32, reflect.Int:
		retCol.sqlType = "INT"
	case reflect.Int64:
		retCol.sqlType = "BIGINT"
	case reflect.Uint8:
		retCol.sqlType = "TINYINT UNSIGNED"
	case reflect.Uint16:
		retCol.sqlType = "SMALLINT UNSIGNED"
	case reflect.Uint32, reflect.Uint:
		retCol.sqlType = "INT UNSIGNED"
	case reflect.Uint64:
		retCol.sqlType = "BIGINT UNSIGNED"
	case reflect.Float32:
		retCol.sqlType = "FLOAT"
	case reflect.Float64:
		retCol.sqlType = "DOUBLE"
	case reflect.Bool:
		retCol.sqlType = "BOOLEAN"
	case reflect.Slice:
		if actualType == reflect.TypeOf([]byte{}) {
			retCol.sqlType = "LONGBLOB"
		} else {
			return nil, fmt.Errorf("unsupported slice type: %s", actualType.String())
		}
	case reflect.Struct:
		if actualType == reflect.TypeOf(time.Time{}) {
			retCol.sqlType = "DATETIME"
		} else {
			return nil, fmt.Errorf("unsupported struct type: %s", actualType.Name())
		}
	default:
		return nil, fmt.Errorf("unsupported field type: %s", actualType.Kind().String())
	}

	// Process tags
	if defaultValue, ok := tag["default"]; ok {
		retCol.defaultString = defaultValue
	} else {
		retCol.defaultString = ""
	}

	if nullable, ok := tag[TAG_NULLABLE]; ok && (nullable == "" || nullable == "true" || nullable == "1") {
		retCol.isNullable = true
	} else if nullable, ok := tag[TAG_NULLABLE]; ok && (nullable == "false" || nullable == "0") {
		retCol.isNullable = false
	}

	if primaryKey, ok := tag[TAG_PRIMARY]; ok && (primaryKey == "" || primaryKey == "true" || primaryKey == "1") {
		retCol.isPrimaryKey = true
		retCol.isNullable = false // Primary keys cannot be null
	} else {
		retCol.isPrimaryKey = false
	}

	if autoIncrement, ok := tag[TAG_AUTO_INCREMENT]; ok && (autoIncrement == "" || autoIncrement == "true" || autoIncrement == "1") {
		retCol.SetAutoIncrement(true)
	} else {
		retCol.SetAutoIncrement(false)
	}

	if unique, ok := tag[TAG_UNIQUE]; ok && (unique == "" || unique == "true" || unique == "1") {
		retCol.isUnique = true
	} else {
		retCol.isUnique = false
	}

	if index, ok := tag[TAG_INDEX]; ok && (index == "" || index == "true" || index == "1") {
		retCol.isIndex = true
	} else {
		retCol.isIndex = false
	}

	if allowZero, ok := tag[TAG_ALLOW_ZERO]; ok && (allowZero == "" || allowZero == "true" || allowZero == "1") {
		retCol.isAllowZero = true
	} else {
		retCol.isAllowZero = false
	}

	retCol.tags = tag
	retCol.columnIndex = -1 // Default value, can be set later if needed

	if name, ok := tag[TAG_NAME]; ok {
		retCol.name = name
	} else {
		retCol.name = columnName
	}

	return &retCol, nil
}

func (mariadb *MariaDBDataBase) DropTableSql(tableName string) string {
	return fmt.Sprintf("DROP TABLE IF EXISTS `%s`;", tableName)
}

func (mariadb *MariaDBDataBase) CanInsert() bool {
	return true
}

func (mariadb *MariaDBDataBase) CanInsertOrUpdate() bool {
	return true
}

func (mariadb *MariaDBDataBase) CanUpdate() bool {
	return true
}

func (mariadb *MariaDBDataBase) CanReturnRowsAffected() bool {
	return true
}

func (mariadb *MariaDBDataBase) CanRenameTable() bool {
	return true
}

func (mariadb *MariaDBDataBase) InsertSqlTemplate() string {
	return "INSERT INTO `{{.TableName}}` ({{.Columns}}) VALUES ({{.Values}});"
}

func (mariadb *MariaDBDataBase) UpdateSqlTemplate() string {
	return "UPDATE `{{.TableName}}` SET {{.Updates}} WHERE {{.Conditions}};"
}

func (mariadb *MariaDBDataBase) CreateIndexSqlTemplate() string {
	return "CREATE INDEX `{{.IndexName}}` ON `{{.TableName}}` ({{.Columns}});"
}

func (mariadb *MariaDBDataBase) DropIndexSqlTemplate() string {
	return "DROP INDEX `{{.IndexName}}` ON `{{.TableName}}`;"
}

func (mariadb *MariaDBDataBase) CreateColumnSqlTemplate() string {
	return "ALTER TABLE `{{.TableName}}` ADD COLUMN `{{.ColumnName}}` {{.ColumnType}};"
}

func (mariadb *MariaDBDataBase) UpdateColumnSqlTemplate() string {
	return "ALTER TABLE `{{.TableName}}` MODIFY COLUMN `{{.ColumnName}}` {{.ColumnType}};"
}

func (mariadb *MariaDBDataBase) GetTableDDL(tableName string) (*Table, error) {
	table := &Table{
		name:        tableName,
		columns:     make([]ColumnInterface, 0),
		indexes:     make([]TableIndex, 0),
		constraints: make([]TableForeignKey, 0),
		db:          mariadb,
	}

	columnMap := make(map[string]*MariaDBColumn)
	columnQuery := `
		SELECT
			COLUMN_NAME,
			DATA_TYPE,
			IS_NULLABLE,
			COLUMN_DEFAULT,
			COLUMN_KEY,
			EXTRA
		FROM
			INFORMATION_SCHEMA.COLUMNS
		WHERE
			TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?
		ORDER BY
			ORDINAL_POSITION;
	`

	rows, err := mariadb.db.Query(columnQuery, tableName)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		var colName, dataType, isNullable, columnKey, extra string
		var defaultValue *string
		if err := rows.Scan(&colName, &dataType, &isNullable, &defaultValue, &columnKey, &extra); err != nil {
			return nil, err
		}

		defaultStr := ""
		if defaultValue != nil {
			defaultStr = *defaultValue
		}

		column := &MariaDBColumn{
			BaseColumn: BaseColumn{
				name:          colName,
				sqlType:       dataType,
				isNullable:    isNullable == "YES",
				defaultString: defaultStr,
				isPrimaryKey:  columnKey == "PRI",
			},
		}
		column.SetAutoIncrement(extra == "auto_increment")
		columnMap[colName] = column
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(columnMap) == 0 {
		return nil, nil // Table doesn't exist
	}

	table.columns = make([]ColumnInterface, 0, len(columnMap))
	for _, col := range columnMap {
		table.columns = append(table.columns, col)
	}

	// Query for indexes
	indexQuery := `
		SELECT
			INDEX_NAME,
			COLUMN_NAME,
			NON_UNIQUE
		FROM
			INFORMATION_SCHEMA.STATISTICS
		WHERE
			TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?
			AND INDEX_NAME != 'PRIMARY'
		ORDER BY
			INDEX_NAME, SEQ_IN_INDEX;
	`

	indexRows, err := mariadb.db.Query(indexQuery, tableName)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = indexRows.Close()
	}()

	indexMap := make(map[string][]string)
	uniqueMap := make(map[string]bool)

	for indexRows.Next() {
		var indexName, columnName string
		var nonUnique int
		if err := indexRows.Scan(&indexName, &columnName, &nonUnique); err != nil {
			return nil, err
		}

		indexMap[indexName] = append(indexMap[indexName], columnName)
		if nonUnique == 0 { // 0 means unique, 1 means non-unique
			uniqueMap[indexName] = true
		}
	}

	if err := indexRows.Err(); err != nil {
		return nil, err
	}

	// Create TableIndex objects
	for indexName, columns := range indexMap {
		isUnique := uniqueMap[indexName]
		index := TableIndex{
			name:     indexName,
			columns:  columns,
			isUnique: isUnique,
		}
		table.indexes = append(table.indexes, index)
	}

	return table, nil
}

func (mariadb *MariaDBDataBase) getTableNames() ([]string, error) {
	query := `
		SELECT TABLE_NAME
		FROM INFORMATION_SCHEMA.TABLES
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_TYPE = 'BASE TABLE';
	`

	rows, err := mariadb.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

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

func (mariadb *MariaDBDataBase) getColumnInfo(tableName string) ([]ColumnInterface, error) {
	query := `
		SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COLUMN_DEFAULT, COLUMN_KEY, EXTRA
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION;
	`

	rows, err := mariadb.db.Query(query, tableName)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var columns []ColumnInterface
	for rows.Next() {
		var colName, dataType, isNullable, columnKey, extra string
		var defaultValue *string
		if err := rows.Scan(&colName, &dataType, &isNullable, &defaultValue, &columnKey, &extra); err != nil {
			return nil, err
		}

		defaultStr := ""
		if defaultValue != nil {
			defaultStr = *defaultValue
		}

		column := &MariaDBColumn{
			BaseColumn: BaseColumn{
				name:          colName,
				sqlType:       dataType,
				isNullable:    isNullable == "YES",
				defaultString: defaultStr,
				isPrimaryKey:  columnKey == "PRI",
			},
		}
		column.SetAutoIncrement(extra == "auto_increment")
		columns = append(columns, column)
	}

	return columns, nil
}
