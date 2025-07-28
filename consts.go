package aaronsql


const (
	DEFAULT_QUOTE_CHAR = "_"
)

// Column tags
const (

	// TAG_IGNORE indicates that the column should be ignored in sql operations
	TAG_IGNORE	   = "ignore"
	// TAG_NAME indicates the name of the column in the database
	TAG_NAME       = "name"
	// TAG_WIDTH indicates the width of the column
	TAG_WIDTH      = "width"
	// TAG_CHARSET indicates the character set of the column
	TAG_CHARSET    = "charset"
	// TAG_PRECISION indicates the precision of the column
	TAG_PRECISION  = "precision"
	// TAG_DEFAULT indicates the default value of the column
	TAG_DEFAULT    = "default"
	// TAG_UNIQUE indicates that the column should be unique
	TAG_UNIQUE     = "unique"
	//TAG_INDEX indicates that the column should be indexed
	TAG_INDEX      = "index"
	// TAG_PRIMARY indicates that the column is a primary key
	TAG_PRIMARY    = "primary"
	// TAG_NULLABLE indicates that the column can be null
	TAG_NULLABLE   = "nullable"
	// TAG_AUTO_INCREMENT indicates that the column is auto-incrementing
	TAG_AUTO_INCREMENT = "auto_increment"
	// TAG_AUTO_VERSION indicates that the column is an auto versioning column
	TAG_AUTO_VERSION = "auto_version"
	// TAG_UPDATED_AT indicates that the column is an updated_at timestamp column
	TAG_UPDATED_AT = "updated_at"
	// TAG_CREATED_AT indicates that the column is a created_at timestamp column
	TAG_CREATED_AT = "created_at"
	// TAG_ALLOW_ZERO indicates that the column allows zero values
	TAG_ALLOW_ZERO = "allow_zero"
	// TAG_EXTRA indicates extra information about the column
	TAG_EXTRA      = "extra"
	// TAG_DEFAULT_PART_QUOTE is used to quote the part in model tag
	TAG_DEFAULT_PART_QUOTE = ";"
	// TAG_DEFAULT_KEY_VALUE_QUOTE is used to separate key and value in model tag
	TAG_DEFAULT_KEY_VALUE_QUOTE = "="
)

// SQL operations
const (
	SQL_AND	  = "AND"
	SQL_OR	  = "OR"
	SQL_NOT	  = "NOT"
	SQL_EQ	  = "="
	SQL_NEQ	  = "!="
	SQL_GT	  = ">"
	SQL_GTE	  = ">="
	SQL_LT	  = "<"
	SQL_LTE	  = "<="
	SQL_LIKE  = "LIKE"
	SQL_IN	  = "IN"
	SQL_NOT_IN = "NOT IN"
	SQL_IS_NULL = "IS NULL"
	SQL_IS_NOT_NULL = "IS NOT NULL"
	SQL_BETWEEN = "BETWEEN"
	SQL_ORDER_BY = "ORDER BY"
)