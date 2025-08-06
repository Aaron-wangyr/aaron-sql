package aaronsql

import (
	"fmt"
	"strconv"
)

type ColumnInterface interface {
	Name() string
	Type() string
	Default() string
	SetDefault(defaultValue string)
	SupportDefault() bool
	Nullable() bool
	SetNullable(nullable bool)
	IsPrimaryKey() bool
	SetPrimaryKey(isPrimaryKey bool)

	IsString() bool
	IsDate() bool
	IsUnique() bool
	IsIndex() bool
	IsText() bool
	IsSearchable() bool
	IsAscii() bool
	IsNumeric() bool
	IsZero(v interface{}) bool
	IsAutoIncrement() bool
	SetAutoIncrement(isAutoIncrement bool)
	AutoIncrementOffset() int64
	SetAutoIncrementOffset(offset int64)
	IsPointer() bool
	AllowZero() bool

	GetWidth() int

	GetStructTags() map[string]string

	ConvertFromStringToSQL(value string) interface{}
	ConvertFromValueToSQL(value interface{}) interface{}

	Extra() string
	DefinitionSQL() string
	GetColumnIndex() int
	SetColumnIndex(index int)

	IsAutoVersion() bool

	IsUpdatedAt() bool
	IsCreatedAt() bool
}

type BaseColumn struct {
	name          string
	oldName       string
	sqlType       string
	defaultString string
	isPointer     bool
	isNullable    bool
	isPrimaryKey  bool
	isUnique      bool
	isIndex       bool
	isAllowZero   bool
	tags          map[string]string
	columnIndex   int
}

// Name returns the column name
func (c *BaseColumn) Name() string {
	return c.name
}

// Type returns the column type
func (c *BaseColumn) Type() string {
	return c.sqlType
}

// Default returns the default value as string
func (c *BaseColumn) Default() string {
	return c.defaultString
}

// SetDefault sets the default value
func (c *BaseColumn) SetDefault(defaultValue string) {
	c.defaultString = defaultValue
}

// SupportDefault returns whether the column supports default values
func (c *BaseColumn) SupportDefault() bool {
	return true
}

// Nullable returns whether the column is nullable
func (c *BaseColumn) Nullable() bool {
	return c.isNullable
}

// SetNullable sets whether the column is nullable
func (c *BaseColumn) SetNullable(nullable bool) {
	c.isNullable = nullable
}

// IsPrimaryKey returns whether the column is a primary key
func (c *BaseColumn) IsPrimaryKey() bool {
	return c.isPrimaryKey
}

// SetPrimaryKey sets whether the column is a primary key
func (c *BaseColumn) SetPrimaryKey(isPrimaryKey bool) {
	c.isPrimaryKey = isPrimaryKey
}

// IsString returns whether the column is a string type
func (c *BaseColumn) IsString() bool {
	return false
}

// IsDate returns whether the column is a date type
func (c *BaseColumn) IsDate() bool {
	return false
}

// IsUnique returns whether the column is unique
func (c *BaseColumn) IsUnique() bool {
	return c.isUnique
}

// IsIndex returns whether the column is indexed
func (c *BaseColumn) IsIndex() bool {
	return c.isIndex
}

// IsText returns whether the column is a text type
func (c *BaseColumn) IsText() bool {
	return false
}

// IsSearchable returns whether the column is searchable
func (c *BaseColumn) IsSearchable() bool {
	return false
}

// IsAscii returns whether the column contains ASCII data
func (c *BaseColumn) IsAscii() bool {
	return false
}

// IsNumeric returns whether the column is numeric
func (c *BaseColumn) IsNumeric() bool {
	return false
}

// IsZero checks if a value is zero for this column type
func (c *BaseColumn) IsZero(v interface{}) bool {
	return false
}

// IsAutoIncrement returns whether the column is auto-incrementing
func (c *BaseColumn) IsAutoIncrement() bool {
	return false
}

// SetAutoIncrement sets whether the column is auto-incrementing
func (c *BaseColumn) SetAutoIncrement(isAutoIncrement bool) {
}

// AutoIncrementOffset returns the auto-increment offset
func (c *BaseColumn) AutoIncrementOffset() int64 {
	return 0
}

// SetAutoIncrementOffset sets the auto-increment offset
func (c *BaseColumn) SetAutoIncrementOffset(offset int64) {
}

// IsPointer returns whether the column is a pointer type
func (c *BaseColumn) IsPointer() bool {
	return c.isPointer
}

// AllowZero returns whether the column allows zero values
func (c *BaseColumn) AllowZero() bool {
	return c.isAllowZero
}

// GetWidth returns the column width
func (c *BaseColumn) GetWidth() int {
	return 0
}

// GetStructTags returns the struct tags for the column
func (c *BaseColumn) GetStructTags() map[string]string {
	return c.tags
}

// ConvertFromStringToSQL converts a string value to SQL format
func (c *BaseColumn) ConvertFromStringToSQL(value string) interface{} {
	return value
}

// ConvertFromValueToSQL converts a value to SQL format
func (c *BaseColumn) ConvertFromValueToSQL(value interface{}) interface{} {
	return value
}

// Extra returns extra column information
func (c *BaseColumn) Extra() string {
	return ""
}

// DefinitionSQL returns the SQL definition for the column
func (c *BaseColumn) DefinitionSQL() string {
	return ""
}

// GetColumnIndex returns the column index
func (c *BaseColumn) GetColumnIndex() int {
	return c.columnIndex
}

// SetColumnIndex sets the column index
func (c *BaseColumn) SetColumnIndex(index int) {
	c.columnIndex = index
}

// IsAutoVersion returns whether the column is an auto-version column
func (c *BaseColumn) IsAutoVersion() bool {
	// Basic implementation - may need to be customized
	return false
}

// IsUpdatedAt returns whether the column is an updated_at timestamp column
func (c *BaseColumn) IsUpdatedAt() bool {
	return false
}

// IsCreatedAt returns whether the column is a created_at timestamp column
func (c *BaseColumn) IsCreatedAt() bool {
	return false
}

func NewBaseColumn(name string, sqltype string, tagmap map[string]string, isPointer bool) BaseColumn {
	var v string
	var ok bool
	if v, ok = tagmap[TAG_NAME]; ok {
		name = v
	}
	defaultStr := ""
	if v, ok = tagmap[TAG_DEFAULT]; ok {
		defaultStr = v
	}
	isNullable := true
	if v, ok = tagmap[TAG_NULLABLE]; ok {
		b, _ := strconv.ParseBool(v)
		isNullable = b
	}
	isPrimaryKey := false
	if v, ok = tagmap[TAG_PRIMARY]; ok {
		if v == "" || v == "true" || v == "1" {
			isPrimaryKey = true
		} else {
			b, _ := strconv.ParseBool(v)
			isPrimaryKey = b
		}
	}
	isUnique := false
	if v, ok = tagmap[TAG_UNIQUE]; ok {
		if v == "" || v == "true" || v == "1" {
			isUnique = true
		} else {
			b, _ := strconv.ParseBool(v)
			isUnique = b
		}
	}
	isIndex := false
	if v, ok = tagmap[TAG_INDEX]; ok {
		if v == "" || v == "true" || v == "1" {
			isIndex = true
		} else {
			b, _ := strconv.ParseBool(v)
			isIndex = b
		}
	}
	if isPrimaryKey {
		// If the column is a primary key, it cannot be nullable
		isNullable = false
	}
	isAllowZero := false
	if v, ok = tagmap[TAG_ALLOW_ZERO]; ok {
		if v == "" || v == "true" || v == "1" {
			isAllowZero = true
		} else {
			b, _ := strconv.ParseBool(v)
			isAllowZero = b
		}
	}
	return BaseColumn{
		name:          name,
		sqlType:       sqltype,
		defaultString: defaultStr,
		isPointer:     isPointer,
		isNullable:    isNullable,
		isPrimaryKey:  isPrimaryKey,
		isUnique:      isUnique,
		isIndex:       isIndex,
		isAllowZero:   isAllowZero,
		tags:          tagmap,
		columnIndex:   -1, // Default index is -1, to be set later
		oldName:       "",
	}
}

type BaseWidthColumn struct {
	BaseColumn
	width int
}

func (c *BaseWidthColumn) ColType() string {
	if c.width > 0 {
		return fmt.Sprintf("%s(%d)", c.Type(), c.width)
	}
	return c.Type()
}

func (c *BaseWidthColumn) GetWidth() int {
	return c.width
}

func NewBaseWidthColumn(name string, sqltype string, tagmap map[string]string, isPointer bool) BaseWidthColumn {
	width := 0
	if v, ok := tagmap[TAG_WIDTH]; ok {
		width, _ = strconv.Atoi(v)
	}
	baseCol := NewBaseColumn(name, sqltype, tagmap, isPointer)
	return BaseWidthColumn{
		BaseColumn: baseCol,
		width:      width,
	}
}
