package aaronsql

type TableForeignKey struct {
	name string
	columns []string
	referencedTable string
	referencedColumns []string
}

func NewTableForeignKey(name string, columns []string, referencedTable string, referencedColumns []string) *TableForeignKey {
	return &TableForeignKey{
		name: name,
		columns: columns,
		referencedTable: referencedTable,
		referencedColumns: referencedColumns,
	}
}
