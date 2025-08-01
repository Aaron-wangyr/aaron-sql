package aaronsql

import "strings"

// Column tags
const (
	// TAG_IGNORE indicates that the column should be ignored in sql operations
	TAG_IGNORE = "ignore"
	// TAG_NAME indicates the name of the column in the database
	TAG_NAME = "name"
	// TAG_WIDTH indicates the width of the column
	TAG_WIDTH = "width"
	// TAG_CHARSET indicates the character set of the column
	TAG_CHARSET = "charset"
	// TAG_PRECISION indicates the precision of the column
	TAG_PRECISION = "precision"
	// TAG_DEFAULT indicates the default value of the column
	TAG_DEFAULT = "default"
	// TAG_UNIQUE indicates that the column should be unique
	TAG_UNIQUE = "unique"
	//TAG_INDEX indicates that the column should be indexed
	TAG_INDEX = "index"
	// TAG_PRIMARY indicates that the column is a primary key
	TAG_PRIMARY = "primary"
	// TAG_NULLABLE indicates that the column can be null
	TAG_NULLABLE = "nullable"
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
	TAG_EXTRA = "extra"
	// TAG_DEFAULT_PART_QUOTE is used to quote the part in model tag
	TAG_DEFAULT_PART_QUOTE = ";"
	// TAG_DEFAULT_KEY_VALUE_QUOTE is used to separate key and value in model tag
	TAG_DEFAULT_KEY_VALUE_QUOTE = ":"
)

var defaultModelDBTagKey = "db"

func SetDefaultModelDBTagKey(key string) {
	if key == "" {
		panic("default model db tag key cannot be empty")
	}
	defaultModelDBTagKey = key
}

func parseTagString(tagStr string) map[string]string {
	if tagStr == "" {
		return nil
	}
	tagMap := make(map[string]string)
	parts := strings.Split(tagStr, TAG_DEFAULT_PART_QUOTE)
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if kv := strings.SplitN(part, TAG_DEFAULT_KEY_VALUE_QUOTE, 2); len(kv) == 2 {
			tagMap[kv[0]] = kv[1]
		} else if len(kv) == 1 {
			tagMap[kv[0]] = ""
		}
	}
	return tagMap
}
