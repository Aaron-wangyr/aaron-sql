package aaronsql

import (
	"database/sql"
	"reflect"
)

type DBName string

const (
	PostgresDB DBName = "postgres"
	MariaDB    DBName = "mariadb"
)

type DBInterface interface {
	// GetName returns the name of the database type.
	Name() DBName
	// GetDB returns the underlying sql.DB instance.
	GetDB() *DataBase
	// GetTables returns the DDL information for all tables in the database.
	GetTables() ([]Table, error)
	GetTableDDL(tableName string) (*Table, error)

	GetCreateTableSQL(tableName string, columns []ColumnInterface) string

	IsSupportForeignKeys() bool
	GetTablesColumns(t TableInterface) ([]ColumnInterface, error)
	GetColumnDefinitionByType(fieldType reflect.Type, columnName string, tag map[string]string, isPointer bool) (ColumnInterface, error)

	DropTableSql(tableName string) string

	CanInsert() bool
	CanInsertOrUpdate() bool
	CanUpdate() bool
	CanReturnRowsAffected() bool
	CanRenameTable() bool

	InsertSqlTemplate() string
	UpdateSqlTemplate() string

	CreateIndexSqlTemplate() string
	DropIndexSqlTemplate() string

	CreateColumnSqlTemplate() string
	UpdateColumnSqlTemplate() string
}

type DataBase struct {
	name DBName
	db   *sql.DB
}

var globalDBInstances = make(map[string]DBInterface)
