package aaronsql

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

// Test struct for sync functionality
type TestUser struct {
	ID        int64     `db:"primary_key:true;auto_increment:true"`
	Name      string    `db:"length:100;nullable:false"`
	Email     string    `db:"length:255;unique:true"`
	Age       *int      `db:"nullable:true"`
	IsActive  bool      `db:"default:true"`
	CreatedAt time.Time `db:"nullable:false"`
}

type TestProduct struct {
	ID          int64   `db:"primary_key:true;auto_increment:true"`
	Name        string  `db:"length:200;nullable:false;index:idx_product_name"`
	Price       float64 `db:"nullable:false"`
	CategoryID  *int64  `db:"nullable:true;index:idx_category_id"`
	Description *string `db:"nullable:true"`
}

const (
	postgresConnStr = "postgres://admin:admin@123@localhost:5432/local-test?sslmode=disable"
	mariadbConnStr  = "admin:admin@123@tcp(localhost:3306)/local-test?parseTime=true"
)

// Helper function to force drop tables in case of test failures
func forceDropTables(dbName string, tableNames ...string) {
	if db, exists := globalDBInstances[dbName]; exists {
		for _, tableName := range tableNames {
			sql := db.DropTableSql(tableName)
			if sql != "" {
				_, _ = db.GetDB().db.Exec(sql)
			}
		}
	}
}

func setupPostgresDB(t *testing.T) (*PostgresDataBase, func()) {
	db, err := sql.Open("postgres", postgresConnStr)
	if err != nil {
		t.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping PostgreSQL: %v", err)
	}

	postgresDB := &PostgresDataBase{
		DataBase: DataBase{
			name: PostgresDB,
			db:   db,
		},
	}

	cleanup := func() {
		// Drop test tables using our Drop method if possible
		if globalDBInstances["postgres_test"] != nil {
			if table, err := NewTableFromStructWithDB(TestUser{}, "test_users", "postgres_test"); err == nil {
				_ = table.Drop() // Ignore errors during cleanup
			}
			if table, err := NewTableFromStructWithDB(TestProduct{}, "test_products", "postgres_test"); err == nil {
				_ = table.Drop() // Ignore errors during cleanup
			}
		}
		// Fallback to direct SQL
		_, _ = db.Exec("DROP TABLE IF EXISTS test_users")
		_, _ = db.Exec("DROP TABLE IF EXISTS test_products")
		_ = db.Close()
		delete(globalDBInstances, "postgres_test")
	}

	return postgresDB, cleanup
}

func setupMariaDB(t *testing.T) (*MariaDBDataBase, func()) {
	db, err := sql.Open("mysql", mariadbConnStr)
	if err != nil {
		t.Fatalf("Failed to connect to MariaDB: %v", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping MariaDB: %v", err)
	}

	mariaDB := &MariaDBDataBase{
		DataBase: DataBase{
			name: MariaDB,
			db:   db,
		},
	}

	cleanup := func() {
		// Drop test tables using our Drop method if possible
		if globalDBInstances["mariadb_test"] != nil {
			if table, err := NewTableFromStructWithDB(TestUser{}, "test_users", "mariadb_test"); err == nil {
				_ = table.Drop() // Ignore errors during cleanup
			}
			if table, err := NewTableFromStructWithDB(TestProduct{}, "test_products", "mariadb_test"); err == nil {
				_ = table.Drop() // Ignore errors during cleanup
			}
		}
		// Fallback to direct SQL
		_, _ = db.Exec("DROP TABLE IF EXISTS test_users")
		_, _ = db.Exec("DROP TABLE IF EXISTS test_products")
		_ = db.Close()
		delete(globalDBInstances, "mariadb_test")
	}

	return mariaDB, cleanup
}

func TestPostgresSyncCreateTable(t *testing.T) {
	db, cleanup := setupPostgresDB(t)
	defer cleanup()

	// Ensure cleanup happens even if test fails
	defer forceDropTables("postgres_test", "test_users", "test_products")

	// Register the database instance
	globalDBInstances["postgres_test"] = db

	// Create table from struct
	table, err := NewTableFromStructWithDB(TestUser{}, "test_users", "postgres_test")
	if err != nil {
		t.Fatalf("Failed to create table from struct: %v", err)
	}

	// Test sync - should create the table
	err = table.Sync()
	if err != nil {
		t.Fatalf("Failed to sync table: %v", err)
	}

	// Verify table exists
	var count int
	err = db.db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'test_users'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check table existence: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected table to exist, but it doesn't")
	}

	// Verify columns exist
	rows, err := db.db.Query(`
		SELECT column_name, data_type, is_nullable 
		FROM information_schema.columns 
		WHERE table_schema = 'public' AND table_name = 'test_users' 
		ORDER BY ordinal_position
	`)
	if err != nil {
		t.Fatalf("Failed to get column info: %v", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	columnCount := 0
	for rows.Next() {
		var colName, dataType, isNullable string
		err := rows.Scan(&colName, &dataType, &isNullable)
		if err != nil {
			t.Fatalf("Failed to scan column info: %v", err)
		}
		columnCount++
		t.Logf("Column: %s, Type: %s, Nullable: %s", colName, dataType, isNullable)
	}

	if columnCount == 0 {
		t.Errorf("Expected columns to exist, but found none")
	}

	// Test sync again - should not fail
	err = table.Sync()
	if err != nil {
		t.Fatalf("Failed to sync existing table: %v", err)
	}
}

func TestMariaDBSyncCreateTable(t *testing.T) {
	db, cleanup := setupMariaDB(t)
	defer cleanup()

	// Ensure cleanup happens even if test fails
	defer forceDropTables("mariadb_test", "test_users", "test_products")

	// Register the database instance
	globalDBInstances["mariadb_test"] = db

	// Create table from struct
	table, err := NewTableFromStructWithDB(TestUser{}, "test_users", "mariadb_test")
	if err != nil {
		t.Fatalf("Failed to create table from struct: %v", err)
	}

	// Test sync - should create the table
	err = table.Sync()
	if err != nil {
		t.Fatalf("Failed to sync table: %v", err)
	}

	// Verify table exists
	var count int
	err = db.db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'local-test' AND table_name = 'test_users'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check table existence: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected table to exist, but it doesn't")
	}

	// Verify columns exist
	rows, err := db.db.Query(`
		SELECT column_name, data_type, is_nullable, column_key, extra
		FROM information_schema.columns 
		WHERE table_schema = 'local-test' AND table_name = 'test_users' 
		ORDER BY ordinal_position
	`)
	if err != nil {
		t.Fatalf("Failed to get column info: %v", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	columnCount := 0
	for rows.Next() {
		var colName, dataType, isNullable, columnKey, extra string
		err := rows.Scan(&colName, &dataType, &isNullable, &columnKey, &extra)
		if err != nil {
			t.Fatalf("Failed to scan column info: %v", err)
		}
		columnCount++
		t.Logf("Column: %s, Type: %s, Nullable: %s, Key: %s, Extra: %s", colName, dataType, isNullable, columnKey, extra)
	}

	if columnCount == 0 {
		t.Errorf("Expected columns to exist, but found none")
	}

	// Test sync again - should not fail
	err = table.Sync()
	if err != nil {
		t.Fatalf("Failed to sync existing table: %v", err)
	}
}

func TestPostgresSyncAddColumn(t *testing.T) {
	db, cleanup := setupPostgresDB(t)
	defer cleanup()

	globalDBInstances["postgres_test"] = db

	// Create initial table with basic structure
	table, err := NewTableFromStructWithDB(TestUser{}, "test_users", "postgres_test")
	if err != nil {
		t.Fatalf("Failed to create table from struct: %v", err)
	}

	err = table.Sync()
	if err != nil {
		t.Fatalf("Failed to sync initial table: %v", err)
	}

	// Define a new struct with an additional column
	type ExtendedUser struct {
		ID        int64      `db:"primary_key:true;auto_increment:true"`
		Name      string     `db:"length:100;nullable:false"`
		Email     string     `db:"length:255;unique:true"`
		Age       *int       `db:"nullable:true"`
		IsActive  bool       `db:"default:true"`
		CreatedAt time.Time  `db:"nullable:false"`
		LastLogin *time.Time `db:"nullable:true"` // New column
	}

	// Create new table definition with additional column
	extendedTable, err := NewTableFromStructWithDB(ExtendedUser{}, "test_users", "postgres_test")
	if err != nil {
		t.Fatalf("Failed to create extended table from struct: %v", err)
	}

	// Sync should add the new column
	err = extendedTable.Sync()
	if err != nil {
		t.Fatalf("Failed to sync extended table: %v", err)
	}

	// Verify new column exists
	var count int
	err = db.db.QueryRow(`
		SELECT COUNT(*) FROM information_schema.columns 
		WHERE table_name = 'test_users' AND column_name = 'lastlogin'
	`).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check new column existence: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected new column 'lastlogin' to exist, but it doesn't")
	}
}

func TestMariaDBSyncAddColumn(t *testing.T) {
	db, cleanup := setupMariaDB(t)
	defer cleanup()

	globalDBInstances["mariadb_test"] = db

	// Create initial table with basic structure
	table, err := NewTableFromStructWithDB(TestUser{}, "test_users", "mariadb_test")
	if err != nil {
		t.Fatalf("Failed to create table from struct: %v", err)
	}

	err = table.Sync()
	if err != nil {
		t.Fatalf("Failed to sync initial table: %v", err)
	}

	// Define a new struct with an additional column
	type ExtendedUser struct {
		ID        int64      `db:"primary_key:true;auto_increment:true"`
		Name      string     `db:"length:100;nullable:false"`
		Email     string     `db:"length:255;unique:true"`
		Age       *int       `db:"nullable:true"`
		IsActive  bool       `db:"default:true"`
		CreatedAt time.Time  `db:"nullable:false"`
		LastLogin *time.Time `db:"nullable:true"` // New column
	}

	// Create new table definition with additional column
	extendedTable, err := NewTableFromStructWithDB(ExtendedUser{}, "test_users", "mariadb_test")
	if err != nil {
		t.Fatalf("Failed to create extended table from struct: %v", err)
	}

	// Sync should add the new column
	err = extendedTable.Sync()
	if err != nil {
		t.Fatalf("Failed to sync extended table: %v", err)
	}

	// Verify new column exists
	var count int
	err = db.db.QueryRow(`
		SELECT COUNT(*) FROM information_schema.columns 
		WHERE table_schema = 'local-test' AND table_name = 'test_users' AND column_name = 'LastLogin'
	`).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check new column existence: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected new column 'LastLogin' to exist, but it doesn't")
	}
}

func TestPostgresSyncWithIndexes(t *testing.T) {
	db, cleanup := setupPostgresDB(t)
	defer cleanup()

	globalDBInstances["postgres_test"] = db

	// Create table with indexes
	table, err := NewTableFromStructWithDB(TestProduct{}, "test_products", "postgres_test")
	if err != nil {
		t.Fatalf("Failed to create table from struct: %v", err)
	}

	err = table.Sync()
	if err != nil {
		t.Fatalf("Failed to sync table with indexes: %v", err)
	}

	// Verify indexes exist
	rows, err := db.db.Query(`
		SELECT indexname, indexdef 
		FROM pg_indexes 
		WHERE tablename = 'test_products'
	`)
	if err != nil {
		t.Fatalf("Failed to get index info: %v", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	indexCount := 0
	for rows.Next() {
		var indexName, indexDef string
		err := rows.Scan(&indexName, &indexDef)
		if err != nil {
			t.Fatalf("Failed to scan index info: %v", err)
		}
		indexCount++
		t.Logf("Index: %s, Definition: %s", indexName, indexDef)
	}

	if indexCount == 0 {
		t.Errorf("Expected indexes to exist, but found none")
	}
}

func TestMariaDBSyncWithIndexes(t *testing.T) {
	db, cleanup := setupMariaDB(t)
	defer cleanup()

	globalDBInstances["mariadb_test"] = db

	// Create table with indexes
	table, err := NewTableFromStructWithDB(TestProduct{}, "test_products", "mariadb_test")
	if err != nil {
		t.Fatalf("Failed to create table from struct: %v", err)
	}

	err = table.Sync()
	if err != nil {
		t.Fatalf("Failed to sync table with indexes: %v", err)
	}

	// Verify indexes exist
	rows, err := db.db.Query(`
		SELECT index_name, column_name, non_unique
		FROM information_schema.statistics 
		WHERE table_schema = 'local-test' AND table_name = 'test_products' AND index_name != 'PRIMARY'
		ORDER BY index_name, seq_in_index
	`)
	if err != nil {
		t.Fatalf("Failed to get index info: %v", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	indexCount := 0
	for rows.Next() {
		var indexName, columnName string
		var nonUnique int
		err := rows.Scan(&indexName, &columnName, &nonUnique)
		if err != nil {
			t.Fatalf("Failed to scan index info: %v", err)
		}
		indexCount++
		unique := "UNIQUE"
		if nonUnique == 1 {
			unique = "NON-UNIQUE"
		}
		t.Logf("Index: %s, Column: %s, Type: %s", indexName, columnName, unique)
	}

	if indexCount == 0 {
		t.Errorf("Expected indexes to exist, but found none")
	}
}

func TestSyncWithBothDatabases(t *testing.T) {
	// Test PostgreSQL
	t.Run("PostgreSQL", func(t *testing.T) {
		db, cleanup := setupPostgresDB(t)
		defer cleanup()

		globalDBInstances["postgres_test"] = db

		table, err := NewTableFromStructWithDB(TestUser{}, "test_users", "postgres_test")
		if err != nil {
			t.Fatalf("Failed to create PostgreSQL table: %v", err)
		}

		err = table.Sync()
		if err != nil {
			t.Fatalf("Failed to sync PostgreSQL table: %v", err)
		}

		// Verify table creation
		var count int
		err = db.db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'test_users'").Scan(&count)
		if err != nil {
			t.Fatalf("Failed to verify PostgreSQL table: %v", err)
		}
		if count != 1 {
			t.Errorf("PostgreSQL table not created")
		}
	})

	// Test MariaDB
	t.Run("MariaDB", func(t *testing.T) {
		db, cleanup := setupMariaDB(t)
		defer cleanup()

		globalDBInstances["mariadb_test"] = db

		table, err := NewTableFromStructWithDB(TestUser{}, "test_users", "mariadb_test")
		if err != nil {
			t.Fatalf("Failed to create MariaDB table: %v", err)
		}

		err = table.Sync()
		if err != nil {
			t.Fatalf("Failed to sync MariaDB table: %v", err)
		}

		// Verify table creation
		var count int
		err = db.db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'local-test' AND table_name = 'test_users'").Scan(&count)
		if err != nil {
			t.Fatalf("Failed to verify MariaDB table: %v", err)
		}
		if count != 1 {
			t.Errorf("MariaDB table not created")
		}
	})
}

func TestPostgresDropTable(t *testing.T) {
	postgresDB, cleanup := setupPostgresDB(t)
	defer cleanup()

	// Register database
	globalDBInstances["postgres_test"] = postgresDB

	// Create table using struct
	table, err := NewTableFromStructWithDB(TestUser{}, "test_drop_table", "postgres_test")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Sync to create the table
	err = table.Sync()
	if err != nil {
		t.Fatalf("Failed to sync table: %v", err)
	}

	// Verify table exists
	var count int
	err = postgresDB.db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'test_drop_table'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check table existence: %v", err)
	}
	if count == 0 {
		t.Fatalf("Table was not created")
	}

	// Drop the table
	err = table.Drop()
	if err != nil {
		t.Fatalf("Failed to drop table: %v", err)
	}

	// Verify table doesn't exist
	err = postgresDB.db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'test_drop_table'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check table existence after drop: %v", err)
	}
	if count > 0 {
		t.Fatalf("Table still exists after drop")
	}
}

func TestMariaDBDropTable(t *testing.T) {
	mariaDB, cleanup := setupMariaDB(t)
	defer cleanup()

	// Register database
	globalDBInstances["mariadb_test"] = mariaDB

	// Create table using struct
	table, err := NewTableFromStructWithDB(TestUser{}, "test_drop_table", "mariadb_test")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Sync to create the table
	err = table.Sync()
	if err != nil {
		t.Fatalf("Failed to sync table: %v", err)
	}

	// Verify table exists
	var count int
	err = mariaDB.db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'test_drop_table'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check table existence: %v", err)
	}
	if count == 0 {
		t.Fatalf("Table was not created")
	}

	// Drop the table
	err = table.Drop()
	if err != nil {
		t.Fatalf("Failed to drop table: %v", err)
	}

	// Verify table doesn't exist
	err = mariaDB.db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'test_drop_table'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check table existence after drop: %v", err)
	}
	if count > 0 {
		t.Fatalf("Table still exists after drop")
	}
}
