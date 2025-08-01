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
	InsertOrUpdate(dst interface{}) error
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

	GetDDL() string
	SyncSql() []string
	Sync()
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
}

func (t *Table) InsertOrUpdate(dst interface{}) error {
	panic("not implemented") // TODO: Implement
}

func (t *Table) Update(dst interface{}, updateFunc func() error) error {
	panic("not implemented") // TODO: Implement
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

func (t *Table) SyncSql() []string {
	return nil
}

func (t *Table) Sync() {
	t.db.getta
}

func (t *Table) DataBase() *DataBase {
	return t.db.GetDB()
}

func (t *Table) Drop() error {
	panic("not implemented") // TODO: Implement
}

func (t *Table) GetExtra() map[string]string {
	panic("not implemented") // TODO: Implement
}

func (t *Table) SetExtra(kvdata map[string]string) {
	panic("not implemented") // TODO: Implement
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
		tags := make(map[string]string)
		field := reflectType.Field(i)
		tagStr := field.Tag.Get(defaultModelDBTagKey)
		if tagStr == "" {
			continue
		}
		tags = parseTagString(tagStr)
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
