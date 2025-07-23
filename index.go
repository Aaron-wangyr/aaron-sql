package aaronsql

import (
	"fmt"
	"sort"
)

type TableIndex struct {
	name    string
	columns []string

	isUnique bool

	table TableInterface
}

func NewTableIndex(table TableInterface, name string, cols []string, unique bool) TableIndex {
	sort.Strings(cols)
	return TableIndex{
		name:     name,
		columns:  cols,
		isUnique: unique,
		table:    table,
	}
}

const IndexLimit = 64

func (i *TableIndex) Name() string {
	if i.name != "" {
		return i.name
	}
	idxName := fmt.Sprintf("idx_%s_%s", i.table.ConstructType().Name(), i.columns[0])
	if len(idxName) > IndexLimit {
		idxName = idxName[:IndexLimit]
	}
	return idxName
}

func (i *TableIndex) clone(table TableInterface) TableIndex {
	return NewTableIndex(table, "", i.columns, i.isUnique)
}

func (i *TableIndex) IsIdentical(columns ...string) bool {
	if len(i.columns) != len(columns) {
		return false
	}
	sort.Strings(columns)
	for j := 0; j < len(i.columns); j++ {
		if i.columns[j] != columns[j] {
			return false
		}
	}
	return true
}

func (i *TableIndex) QuotedColumns(quoteStr string) []string {
	ret := make([]string, len(i.columns))
	for j := 0; j < len(ret); j++ {
		ret[j] = fmt.Sprintf("%s%s%s", quoteStr, i.columns[j], quoteStr)
	}
	return ret
}
