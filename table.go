package aaronsql

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// TableInterface respresents the interface for table operations in the database.
type TableInterface interface {
	// Insert inserts a new record into the table.
	Insert(dst interface{}) error
	Update(dst interface{}, updateFunc func() error) error

	ConstructType() reflect.Type
	Column(name string) ColumnInterface
	Name() string
	Columns() []ColumnInterface
	PrimaryColumns() []ColumnInterface
	Indexes() []TableIndex
	Instance() *Table
	DropForeignKeySql() string
	AddIndex(unique bool, cols ...string) bool

	DataBase() *DataBase
	Drop() error
	GetExtra() map[string]string
	SetExtra(kvdata map[string]string)

	Sync() error
}

type Table struct {
	structType  reflect.Type
	name        string
	columns     []ColumnInterface
	indexes     []TableIndex
	constraints []TableForeignKey

	extraOptions map[string]string

	db DBInterface
}

func (t *Table) Name() string {
	return t.name
}

// Insert inserts a new record into the table.
func (t *Table) Insert(dst interface{}) error {
	if !t.db.CanInsert() {
		return fmt.Errorf("insert operation is not supported for database: %s", t.db.Name())
	}

	// Use reflection to get values from the struct
	reflectValue := reflect.ValueOf(dst)
	if reflectValue.Kind() == reflect.Ptr {
		reflectValue = reflectValue.Elem()
	}
	if reflectValue.Kind() != reflect.Struct {
		return fmt.Errorf("expected a struct or pointer to struct, got: %s", reflectValue.Kind().String())
	}

	// Build column names and values for INSERT
	var columnNames []string
	var placeholders []string
	var values []interface{}
	placeholderIndex := 1

	for _, col := range t.columns {
		// Skip auto-increment columns for insert
		if col.IsAutoIncrement() {
			continue
		}

		// Get field value by column name
		fieldValue := reflectValue.FieldByName(col.Name())
		if !fieldValue.IsValid() {
			// Try to find field by struct tag name
			found := false
			for i := 0; i < reflectValue.NumField(); i++ {
				field := reflectValue.Type().Field(i)
				tagStr := field.Tag.Get(defaultModelDBTagKey)
				if tagStr != "" {
					tags := parseTagString(tagStr)
					if nameTag, ok := tags[TAG_NAME]; ok && nameTag == col.Name() {
						fieldValue = reflectValue.Field(i)
						found = true
						break
					}
				}
				// If no name tag, check if field name matches column name
				if !found && field.Name == col.Name() {
					fieldValue = reflectValue.Field(i)
					found = true
					break
				}
			}
			if !found {
				continue // Skip if field not found
			}
		}

		columnNames = append(columnNames, col.Name())

		// Handle different database placeholder styles
		if t.db.Name() == PostgresDB {
			placeholders = append(placeholders, fmt.Sprintf("$%d", placeholderIndex))
			placeholderIndex++
		} else {
			placeholders = append(placeholders, "?")
		}

		// Convert value using column's conversion method
		sqlValue := col.ConvertFromValueToSQL(fieldValue.Interface())
		values = append(values, sqlValue)
	}

	if len(columnNames) == 0 {
		return fmt.Errorf("no columns to insert")
	}

	// Build and execute INSERT SQL
	insertSQL := t.db.InsertSqlTemplate()
	insertSQL = strings.ReplaceAll(insertSQL, "{{.TableName}}", t.name)
	insertSQL = strings.ReplaceAll(insertSQL, "{{.Columns}}", strings.Join(columnNames, ", "))
	insertSQL = strings.ReplaceAll(insertSQL, "{{.Values}}", strings.Join(placeholders, ", "))

	_, err := t.db.GetDB().db.Exec(insertSQL, values...)
	if err != nil {
		return fmt.Errorf("failed to insert into table %s: %w", t.name, err)
	}

	return nil
}

func (t *Table) Update(dst interface{}, updateFunc func() error) error {
	if !t.db.CanUpdate() {
		return fmt.Errorf("update operation is not supported for database: %s", t.db.Name())
	}

	// Execute the update function first if provided
	if updateFunc != nil {
		if err := updateFunc(); err != nil {
			return fmt.Errorf("update function failed: %w", err)
		}
	}

	// Use reflection to get values from the struct
	reflectValue := reflect.ValueOf(dst)
	if reflectValue.Kind() == reflect.Ptr {
		reflectValue = reflectValue.Elem()
	}
	if reflectValue.Kind() != reflect.Struct {
		return fmt.Errorf("expected a struct or pointer to struct, got: %s", reflectValue.Kind().String())
	}

	// Build update clauses and where conditions
	var updateClauses []string
	var whereConditions []string
	var values []interface{}
	placeholderIndex := 1

	// Get primary key columns for WHERE clause
	primaryCols := t.PrimaryColumns()
	if len(primaryCols) == 0 {
		return fmt.Errorf("no primary key columns found for update operation")
	}

	// Build SET clauses for non-primary key columns
	for _, col := range t.columns {
		if col.IsPrimaryKey() || col.IsAutoIncrement() {
			continue // Skip primary keys and auto-increment columns in SET clause
		}

		// Get field value by column name
		fieldValue := reflectValue.FieldByName(col.Name())
		if !fieldValue.IsValid() {
			// Try to find field by struct tag name
			found := false
			for i := 0; i < reflectValue.NumField(); i++ {
				field := reflectValue.Type().Field(i)
				tagStr := field.Tag.Get(defaultModelDBTagKey)
				if tagStr != "" {
					tags := parseTagString(tagStr)
					if nameTag, ok := tags[TAG_NAME]; ok && nameTag == col.Name() {
						fieldValue = reflectValue.Field(i)
						found = true
						break
					}
				}
				// If no name tag, check if field name matches column name
				if !found && field.Name == col.Name() {
					fieldValue = reflectValue.Field(i)
					found = true
					break
				}
			}
			if !found {
				continue // Skip if field not found
			}
		}

		// Handle different database placeholder styles
		if t.db.Name() == PostgresDB {
			updateClauses = append(updateClauses, fmt.Sprintf("%s = $%d", col.Name(), placeholderIndex))
			placeholderIndex++
		} else {
			updateClauses = append(updateClauses, fmt.Sprintf("%s = ?", col.Name()))
		}

		// Convert value using column's conversion method
		sqlValue := col.ConvertFromValueToSQL(fieldValue.Interface())
		values = append(values, sqlValue)
	}

	// Build WHERE clause using primary key columns
	for _, col := range primaryCols {
		// Get field value by column name
		fieldValue := reflectValue.FieldByName(col.Name())
		if !fieldValue.IsValid() {
			// Try to find field by struct tag name
			found := false
			for i := 0; i < reflectValue.NumField(); i++ {
				field := reflectValue.Type().Field(i)
				tagStr := field.Tag.Get(defaultModelDBTagKey)
				if tagStr != "" {
					tags := parseTagString(tagStr)
					if nameTag, ok := tags[TAG_NAME]; ok && nameTag == col.Name() {
						fieldValue = reflectValue.Field(i)
						found = true
						break
					}
				}
				// If no name tag, check if field name matches column name
				if !found && field.Name == col.Name() {
					fieldValue = reflectValue.Field(i)
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("primary key field %s not found in struct", col.Name())
			}
		}

		// Handle different database placeholder styles
		if t.db.Name() == PostgresDB {
			whereConditions = append(whereConditions, fmt.Sprintf("%s = $%d", col.Name(), placeholderIndex))
			placeholderIndex++
		} else {
			whereConditions = append(whereConditions, fmt.Sprintf("%s = ?", col.Name()))
		}

		// Convert value using column's conversion method
		sqlValue := col.ConvertFromValueToSQL(fieldValue.Interface())
		values = append(values, sqlValue)
	}

	if len(updateClauses) == 0 {
		return fmt.Errorf("no columns to update")
	}

	// Build and execute UPDATE SQL
	updateSQL := t.db.UpdateSqlTemplate()
	updateSQL = strings.ReplaceAll(updateSQL, "{{.TableName}}", t.name)
	updateSQL = strings.ReplaceAll(updateSQL, "{{.Updates}}", strings.Join(updateClauses, ", "))
	updateSQL = strings.ReplaceAll(updateSQL, "{{.Conditions}}", strings.Join(whereConditions, " AND "))

	result, err := t.db.GetDB().db.Exec(updateSQL, values...)
	if err != nil {
		return fmt.Errorf("failed to update table %s: %w", t.name, err)
	}

	// Check if any rows were affected
	if t.db.CanReturnRowsAffected() {
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}
		if rowsAffected == 0 {
			return fmt.Errorf("no rows were updated")
		}
	}

	return nil
}

func (t *Table) ConstructType() reflect.Type {
	return t.structType
}

func (t *Table) Column(name string) ColumnInterface {
	for _, col := range t.columns {
		if col.Name() == name {
			return col
		}
	}
	return nil // or return an error if preferred
}

func (t *Table) Columns() []ColumnInterface {
	return t.columns
}

func (t *Table) PrimaryColumns() []ColumnInterface {
	primaryCols := make([]ColumnInterface, 0)
	for _, col := range t.columns {
		if col.IsPrimaryKey() {
			primaryCols = append(primaryCols, col)
		}
	}
	return primaryCols
}

func (t *Table) Indexes() []TableIndex {
	return t.indexes
}

func (t *Table) Instance() *Table {
	return t
}

func (t *Table) DropForeignKeySql() string {
	return ""
}

// Sync synchronizes the table structure by Table.Only do the create or update operation, non destructive.
func (t *Table) Sync() error {
	existTable, err := t.db.GetTableDDL(t.name)
	if err != nil {
		return fmt.Errorf("failed to get DDL for table %s: %w", t.name, err)
	}
	if existTable == nil {
		// Table does not exist, create it
		createSQL := t.db.GetCreateTableSQL(t.name, t.columns)
		if createSQL == "" {
			return fmt.Errorf("failed to generate CREATE TABLE SQL for table %s", t.name)
		}

		// Execute the CREATE TABLE statement
		_, err := t.db.GetDB().db.Exec(createSQL)
		if err != nil {
			return fmt.Errorf("failed to create table %s: %w", t.name, err)
		}

		// Create indexes if any
		createIndexSQLTemplate := t.db.CreateIndexSqlTemplate()
		for _, index := range t.indexes {
			var indexSQL string
			if index.IsUnique() {
				// For unique indexes, we need to use CREATE UNIQUE INDEX
				indexSQL = strings.ReplaceAll(createIndexSQLTemplate, "CREATE INDEX", "CREATE UNIQUE INDEX")
			} else {
				indexSQL = createIndexSQLTemplate
			}
			indexSQL = strings.ReplaceAll(indexSQL, "{{.IndexName}}", index.Name())
			indexSQL = strings.ReplaceAll(indexSQL, "{{.TableName}}", t.name)
			indexSQL = strings.ReplaceAll(indexSQL, "{{.Columns}}", strings.Join(index.columns, ", "))
			if indexSQL != "" {
				_, err := t.db.GetDB().db.Exec(indexSQL)
				if err != nil {
					return fmt.Errorf("failed to create index %s for table %s: %w", index.Name(), t.name, err)
				}
			}
		}
	} else {
		// Table exists, check for column differences and add missing columns
		existingCols := existTable.Columns()
		existingColNames := make(map[string]bool)
		for _, col := range existingCols {
			existingColNames[col.Name()] = true
		}

		// Add missing columns and update existing ones if they differ
		for _, newCol := range t.columns {
			colExists := false
			var existingCol ColumnInterface

			// Find existing column (case-insensitive for PostgreSQL, case-sensitive for others)
			for _, col := range existingCols {
				if t.db.Name() == PostgresDB {
					// PostgreSQL is case-insensitive, compare lowercase
					if strings.EqualFold(col.Name(), newCol.Name()) {
						colExists = true
						existingCol = col
						break
					}
				} else {
					// Other databases (MariaDB) are case-sensitive
					if col.Name() == newCol.Name() {
						colExists = true
						existingCol = col
						break
					}
				}
			}

			if !colExists {
				// Column doesn't exist, add it
				colSQL := t.db.CreateColumnSqlTemplate()
				colSQL = strings.ReplaceAll(colSQL, "{{.ColumnName}}", newCol.Name())
				colSQL = strings.ReplaceAll(colSQL, "{{.ColumnType}}", newCol.Type())
				colSQL = strings.ReplaceAll(colSQL, "{{.TableName}}", t.name)
				if colSQL != "" {
					_, err := t.db.GetDB().db.Exec(colSQL)
					if err != nil {
						return fmt.Errorf("failed to add column %s to table %s: %w", newCol.Name(), t.name, err)
					}
				}
			} else {
				// Column exists, check if it needs to be updated
				if existingCol != nil {
					// Compare column definitions
					if existingCol.Type() != newCol.Type() ||
						existingCol.Nullable() != newCol.Nullable() ||
						existingCol.Default() != newCol.Default() {
						// Column definition differs, update it
						updateSQL := t.db.UpdateColumnSqlTemplate()
						updateSQL = strings.ReplaceAll(updateSQL, "{{.ColumnName}}", newCol.Name())
						updateSQL = strings.ReplaceAll(updateSQL, "{{.ColumnType}}", newCol.Type())
						updateSQL = strings.ReplaceAll(updateSQL, "{{.TableName}}", t.name)

						if updateSQL != "" {
							_, err := t.db.GetDB().db.Exec(updateSQL)
							if err != nil {
								return fmt.Errorf("failed to update column %s in table %s: %w", newCol.Name(), t.name, err)
							}
						}
					}
				}
			} // Create missing indexes and update existing ones if they differ
			existingIndexes := existTable.Indexes()
			existingIndexMap := make(map[string]TableIndex)
			for _, idx := range existingIndexes {
				existingIndexMap[idx.Name()] = idx
			}

			for _, newIndex := range t.indexes {
				if existingIdx, exists := existingIndexMap[newIndex.Name()]; !exists {
					// Index doesn't exist, create it
					var indexSQL string
					if newIndex.IsUnique() {
						// For unique indexes, we need to use CREATE UNIQUE INDEX
						indexSQL = strings.ReplaceAll(t.db.CreateIndexSqlTemplate(), "CREATE INDEX", "CREATE UNIQUE INDEX")
					} else {
						indexSQL = t.db.CreateIndexSqlTemplate()
					}
					indexSQL = strings.ReplaceAll(indexSQL, "{{.IndexName}}", newIndex.Name())
					indexSQL = strings.ReplaceAll(indexSQL, "{{.TableName}}", t.name)
					indexSQL = strings.ReplaceAll(indexSQL, "{{.Columns}}", strings.Join(newIndex.columns, ", "))
					if indexSQL != "" {
						_, err := t.db.GetDB().db.Exec(indexSQL)
						if err != nil {
							return fmt.Errorf("failed to create index %s for table %s: %w", newIndex.Name(), t.name, err)
						}
					}
				} else {
					// Index exists, check if it matches the current definition
					if !existingIdx.IsIdentical(newIndex.columns...) || existingIdx.isUnique != newIndex.isUnique {
						// Index definition differs, drop and recreate it
						dropIndexSQL := t.db.DropIndexSqlTemplate()
						dropIndexSQL = strings.ReplaceAll(dropIndexSQL, "{{.IndexName}}", existingIdx.Name())
						dropIndexSQL = strings.ReplaceAll(dropIndexSQL, "{{.TableName}}", t.name)

						if dropIndexSQL != "" {
							_, err := t.db.GetDB().db.Exec(dropIndexSQL)
							if err != nil {
								return fmt.Errorf("failed to drop index %s for table %s: %w", existingIdx.Name(), t.name, err)
							}
						}

						// Recreate the index with new definition
						var createIndexSQL string
						if newIndex.IsUnique() {
							// For unique indexes, we need to use CREATE UNIQUE INDEX
							createIndexSQL = strings.ReplaceAll(t.db.CreateIndexSqlTemplate(), "CREATE INDEX", "CREATE UNIQUE INDEX")
						} else {
							createIndexSQL = t.db.CreateIndexSqlTemplate()
						}
						createIndexSQL = strings.ReplaceAll(createIndexSQL, "{{.IndexName}}", newIndex.Name())
						createIndexSQL = strings.ReplaceAll(createIndexSQL, "{{.TableName}}", t.name)
						createIndexSQL = strings.ReplaceAll(createIndexSQL, "{{.Columns}}", strings.Join(newIndex.columns, ", "))

						if createIndexSQL != "" {
							_, err := t.db.GetDB().db.Exec(createIndexSQL)
							if err != nil {
								return fmt.Errorf("failed to recreate index %s for table %s: %w", newIndex.Name(), t.name, err)
							}
						}
					}
				}
			}
		}
	}

	return nil
}

func (t *Table) DataBase() *DataBase {
	return t.db.GetDB()
}

func (t *Table) Drop() error {
	dropSQL := t.db.DropTableSql(t.name)
	if dropSQL == "" {
		return fmt.Errorf("drop table SQL not supported for database: %s", t.db.Name())
	}

	_, err := t.db.GetDB().db.Exec(dropSQL)
	if err != nil {
		return fmt.Errorf("failed to drop table %s: %w", t.name, err)
	}

	return nil
}

func (t *Table) GetExtra() map[string]string {
	if t.extraOptions == nil {
		t.extraOptions = make(map[string]string)
	}
	// Return a copy to prevent external modification
	result := make(map[string]string)
	for k, v := range t.extraOptions {
		result[k] = v
	}
	return result
}

func (t *Table) SetExtra(kvdata map[string]string) {
	if t.extraOptions == nil {
		t.extraOptions = make(map[string]string)
	}
	// Copy the provided data to prevent external modification
	for k, v := range kvdata {
		t.extraOptions[k] = v
	}
}

func NewTableFromStructWithDB(s interface{}, name string, dbName string) (*Table, error) {
	dbRefer := globalDBInstances[dbName]
	if dbRefer == nil {
		return nil, fmt.Errorf("database instance for %s not found", dbName)
	}
	reflectType := reflect.TypeOf(s)
	if reflectType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected a struct type, got: %s", reflectType.Kind().String())
	}
	// get s tags
	cols := make([]ColumnInterface, 0)
	for i := 0; i < reflectType.NumField(); i++ {
		field := reflectType.Field(i)
		tagStr := field.Tag.Get(defaultModelDBTagKey)
		if tagStr == "" {
			continue
		}
		tags := parseTagString(tagStr)
		if _, ok := tags[TAG_IGNORE]; ok {
			continue // skip fields with ignore tag
		}
		// Create a new column based on the field information
		col, err := globalDBInstances[dbName].GetColumnDefinitionByType(field.Type, field.Name, tags, field.Type.Kind() == reflect.Ptr)
		if err != nil {
			return nil, fmt.Errorf("failed to create column for field %s: %w", field.Name, err)
		}
		cols = append(cols, col)
	}

	table := &Table{
		structType:   reflect.TypeOf(s),
		name:         name,
		columns:      cols,
		extraOptions: make(map[string]string),
		db:           dbRefer,
	}
	if err := table.constructIndex(); err != nil {
		return nil, fmt.Errorf("failed to construct indexes for table %s: %w", name, err)
	}
	if err := table.constructConstraints(); err != nil {
		return nil, fmt.Errorf("failed to construct constraints for table %s: %w", name, err)
	}
	return table, nil
}

func (table *Table) addIndexWithName(name string, unique bool, cols ...string) bool {
	for i := 0; i < len(table.indexes); i++ {
		if table.indexes[i].IsIdentical(cols...) {
			return false
		}
	}
	idx := NewTableIndex(table, name, cols, unique)
	table.indexes = append(table.indexes, idx)
	return true
}

func (ts *Table) AddIndex(unique bool, cols ...string) bool {
	return ts.addIndexWithName("", unique, cols...)
}

// constructIndex constructs the indexes for the table based on the columns tags.
// eg: db:"index:idx_name,unique".if two columns have the same index name, it's a composite index.
// composite index will use priority to determine the index order. eg: db:"index:idx_name,priority:1".
// support multiple indexes on the same column, but must with different index type. eg: db:"index:idx_name,priority:1;index:idx_name2,unique".
func (t *Table) constructIndex() error {
	indexs := make([]TableIndex, 0)
	indexNameAndColsAndPriority := make(map[string]map[string]int)
	for _, col := range t.columns {
		tags := col.GetStructTags()
		if indexTag, ok := tags[TAG_INDEX]; ok {
			// If the index tag is present, we need to parse it
			indexParts := strings.Split(indexTag, ",")
			idxName := ""
			priority := 0
			for _, part := range indexParts {
				if strings.HasPrefix(part, "priority:") {
					// Extract priority value
					priorityStr := strings.TrimPrefix(part, "priority:")
					var err error
					priority, err = strconv.Atoi(priorityStr)
					if err != nil {
						return fmt.Errorf("invalid priority value in index tag: %s", part)
					}
				} else {
					// This is the index name
					idxName = part
				}
			}
			if idxName == "" {
				idxName = col.Name() + "_index" // Default index name if not specified
			}
			indexNameAndColsAndPriority[idxName] = map[string]int{col.Name(): priority}
		}
		if uniqueTag, ok := tags[TAG_UNIQUE]; ok {
			if uniqueTag == "true" || uniqueTag == "1" {
				indexs = append(indexs, NewTableIndex(t, col.Name()+"_unique", []string{col.Name()}, true))
			}
		}
	}
	if len(indexNameAndColsAndPriority) != 0 {
		for idxName, colMap := range indexNameAndColsAndPriority {
			cols := make([]string, 0, len(colMap))
			for colName := range colMap {
				cols = append(cols, colName)
			}
			for i := 0; i < len(cols); i++ {
				for j := i + 1; j < len(cols); j++ {
					if colMap[cols[i]] < colMap[cols[j]] {
						// Swap to maintain priority order
						cols[i], cols[j] = cols[j], cols[i]
					}
				}
			}
			indexs = append(indexs, NewTableIndex(t, idxName, cols, false))
		}
	}
	t.indexes = indexs
	return nil
}

func (t *Table) constructConstraints() error {
	// This function is a placeholder for future implementation of foreign key constraints
	// Currently, it does nothing but can be extended later
	return nil
}
