package aaronsql

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

// TestDebugUser for table operations tests
type TestDebugUser struct {
	ID        int64     `db:"name:id;primary"`
	Name      string    `db:"name:name"`
	Email     string    `db:"name:email;unique"`
	Age       int       `db:"name:age;nullable"`
	IsActive  bool      `db:"name:isactive"`
	CreatedAt time.Time `db:"name:createdat;created_at"`
}

func TestInsertOrUpdateDebug(t *testing.T) {
	// Setup PostgreSQL table for testing
	db, err := sql.Open("postgres", postgresConnStr)
	if err != nil {
		t.Skipf("PostgreSQL not available: %v", err)
	}
	defer db.Close()

	dbInstance := &PostgresDataBase{
		DataBase: DataBase{
			name: PostgresDB,
			db:   db,
		},
	}
	globalDBInstances["test_debug"] = dbInstance
	defer delete(globalDBInstances, "test_debug")

	table, err := NewTableFromStructWithDB(TestDebugUser{}, "test_debug_table", "test_debug")
	if err != nil {
		t.Fatalf("Failed to create table from struct: %v", err)
	}

	// Create the table
	err = table.Sync()
	if err != nil {
		t.Fatalf("Failed to sync table: %v", err)
	}
	defer table.Drop()

	// Debug: Check primary columns
	primaryCols := table.PrimaryColumns()
	t.Logf("Primary columns count: %d", len(primaryCols))
	for _, col := range primaryCols {
		t.Logf("Primary column: %s, IsPrimaryKey: %t", col.Name(), col.IsPrimaryKey())
	}

	// Debug: Check all columns and their tags
	allCols := table.Columns()
	t.Logf("All columns count: %d", len(allCols))
	for _, col := range allCols {
		tags := col.GetStructTags()
		t.Logf("Column: %s, IsPrimaryKey: %t, Tags: %+v", col.Name(), col.IsPrimaryKey(), tags)
	}

	user := TestDebugUser{
		ID:        1,
		Name:      "John Doe",
		Email:     "john@example.com",
		Age:       30,
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	err = table.InsertOrUpdate(&user)
	if err != nil {
		t.Fatalf("InsertOrUpdate failed: %v", err)
	}
}
