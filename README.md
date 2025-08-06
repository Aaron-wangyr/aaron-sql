# AaronSQL - Database Schema Synchronization Library

AaronSQL is a Go library that provides database schema synchronization functionality for PostgreSQL and MariaDB. It allows you to define database schemas using Go structs and automatically sync them with your database.

## Features

### Supported Databases
- **PostgreSQL** - Full support for table creation, column management, and index operations
- **MariaDB** - Complete implementation with MariaDB-specific optimizations

### Core Functionality
- **Table Synchronization** - Automatically create and update table schemas based on Go structs
- **Column Management** - Add, modify, and synchronize table columns
- **Index Support** - Create and manage database indexes including unique indexes
- **Drop Tables** - Safe table deletion functionality
- **Type Mapping** - Automatic Go type to SQL type conversion

## Usage

### Basic Table Creation and Sync

```go
package main

import (
    "database/sql"
    "time"
    _ "github.com/lib/pq"           // PostgreSQL driver
    _ "github.com/go-sql-driver/mysql" // MariaDB driver
)

// Define your struct with database tags
type User struct {
    ID        int64     `db:"primary_key:true;auto_increment:true"`
    Name      string    `db:"length:100;nullable:false"`
    Email     string    `db:"length:255;unique:true"`
    Age       *int      `db:"nullable:true"`
    IsActive  bool      `db:"default:true"`
    CreatedAt time.Time `db:"nullable:false"`
}

func main() {
    // Connect to PostgreSQL
    db, err := sql.Open("postgres", "postgres://user:pass@localhost/dbname?sslmode=disable")
    if err != nil {
        panic(err)
    }
    defer db.Close()

    // Create database instance
    pgDB := &PostgresDataBase{
        DataBase: DataBase{
            name: PostgresDB,
            db:   db,
        },
    }

    // Register database
    globalDBInstances["my_db"] = pgDB

    // Create table from struct
    table, err := NewTableFromStructWithDB(User{}, "users", "my_db")
    if err != nil {
        panic(err)
    }

    // Sync the table - creates or updates schema
    err = table.Sync()
    if err != nil {
        panic(err)
    }

    // Drop table when needed
    err = table.Drop()
    if err != nil {
        panic(err)
    }
}
```

### Struct Tags

The library uses struct tags to define database schema properties:

- `primary_key:true` - Mark field as primary key
- `auto_increment:true` - Enable auto increment
- `length:255` - Set column length for strings
- `nullable:true/false` - Control NULL constraints
- `unique:true` - Create unique constraint
- `default:value` - Set default value
- `index:index_name` - Create index on field

### Tag Format
Tags use semicolon (`;`) separation:
```go
Field string `db:"length:100;nullable:false;unique:true"`
```

## Database Support

### PostgreSQL Features
- Complete DDL operations
- Information schema queries
- Index management with BTREE support
- Case-insensitive column handling
- Proper NULL value handling

### MariaDB Features
- Full MySQL/MariaDB compatibility
- Optimized type mappings
- Unique index detection
- AUTO_INCREMENT support
- Boolean type mapping to TINYINT

## Testing

The library includes comprehensive test coverage:

```bash
# Run all tests
go test -v

# Test specific database
go test -v -run "Postgres"
go test -v -run "MariaDB"

# Test specific functionality
go test -v -run "Sync"
go test -v -run "Drop"
```

### Test Requirements
- PostgreSQL server running on localhost:5432
- MariaDB server running on localhost:3306
- Test database: `local-test`
- Credentials: `admin:admin@123`

## Type Mappings

### Go to PostgreSQL
- `string` → `TEXT` (or `VARCHAR(n)` with length tag)
- `int`, `int32` → `INTEGER`
- `int64` → `BIGINT`
- `float64` → `DOUBLE PRECISION`
- `bool` → `BOOLEAN`
- `time.Time` → `TIMESTAMP WITH TIME ZONE`
- `*int`, `*string`, etc. → Nullable versions

### Go to MariaDB
- `string` → `TEXT` (or `VARCHAR(n)` with length tag)
- `int`, `int32` → `INT`
- `int64` → `BIGINT`
- `float64` → `DOUBLE`
- `bool` → `BOOLEAN` (stored as TINYINT)
- `time.Time` → `DATETIME`
- `*int`, `*string`, etc. → Nullable versions

## Architecture

The library follows an interface-based design:

- `DBInterface` - Core database operations interface
- `ColumnInterface` - Column definition interface
- `Table` - Table management and synchronization
- `TableIndex` - Index definition and management

### Database Implementations
- `PostgresDataBase` - PostgreSQL implementation
- `MariaDBDataBase` - MariaDB implementation

Both implementations provide:
- Schema introspection
- DDL generation
- Type conversion
- Index management

## Performance Considerations

- Index detection optimized to prevent duplicate creation
- Batch operations for schema changes
- Efficient schema comparison algorithms
- Minimal database round trips during sync

## Error Handling

The library provides detailed error messages for:
- Connection failures
- Schema validation errors
- Type conversion issues
- DDL execution problems
- Index conflicts

## Contributing

When contributing:
1. Add tests for new functionality
2. Follow Go naming conventions
3. Update documentation
4. Ensure both PostgreSQL and MariaDB compatibility

## License

This project is licensed under the MIT License.
