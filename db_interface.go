package aaronsql

import (
	"database/sql"
	"reflect"
)

type DBName string

const (
	PostgresDB DBName = "postgres"
)

type DBInterface interface {
	// GetName returns the name of the database type.
	Name() DBName
	// GetTablesSQL returns the SQL query to retrieve the list of tables in the database.
	GetTablesSQL() string

	GetCreateTableSQL(tableName string, columns []ColumnInterface) string
	IsSupportForeignKeys() bool
	GetTablesColumns(t TableInterface) ([]ColumnInterface, error)
	GetColumnDefinitionByType(table *Table, fieldType reflect.Type, columnName string, tag map[string]string, isPointer bool) (ColumnInterface, error)

	DropTableSql(tableName string) string

	CanInsert() bool
	CanInsertOrUpdate() bool
	CanUpdate() bool
	
}

type DataBase struct {
	name DBName
	db   *sql.DB
}

var globalDBInstances = make(map[string]DBInterface)

type PostgresDataBase struct {
	schema string
	DataBase
}

func (db *PostgresDataBase) Name() DBName {
	return PostgresDB
}
func (db *PostgresDataBase) GetTablesSQL() string {
	// TODO: Implement the SQL query to retrieve tables for Postgres
	return ""
}

func (db *PostgresDataBase) GetCreateTableSQL(tableName string, columns []ColumnInterface) string {
	// TODO: Implement the SQL query to create a table for Postgres
	return ""
}
