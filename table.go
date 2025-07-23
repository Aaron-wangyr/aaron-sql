package aaronsql

import (
	"reflect"
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
	SyncSql() []string
	Sync()
	DataBase() *DataBase
	Drop() error
	GetExtra() map[string]string
	SetExtra(kvdata map[string]string)
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
	panic("not implemented") // TODO: Implement
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
	panic("not implemented") // TODO: Implement
}

func (t *Table) DataBase() *DataBase {
	return &t.db.DataBase
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



func NewTableFromStructWithDB(s interface{}, name string, dbName string) *Table {
	dbRefer:= globalDBInstances[dbName]
	if dbRefer == nil {
		panic("Database instance not found for name: " + dbName)
	}
	reflectType := reflect.TypeOf(s)
	if reflectType.Kind() != reflect.Struct {
		panic("Expected a struct type, got: " + reflectType.Kind().String())
	}

	table := &Table{
		structType:  reflect.TypeOf(s),
		name:        name,
		columns:     make([]ColumnInterface, 0),
		indexes:     make([]TableIndex, 0),
		constraints: make([]TableForeignKey, 0),
		extraOptions: make(map[string]string),
		db:          dbRefer,
	}
	return table
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
