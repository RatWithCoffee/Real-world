package utils

import (
	"database/sql"
	"fmt"
)

type ListOfColumns map[string]struct{}

func GetDbTablesCols(db *sql.DB) (map[string]ListOfColumns, error) {
	tableNames, err := getTablesNames(db)
	if err != nil {
		return nil, fmt.Errorf("getDBInfo : %v", err)
	}
	dbTablesCols := make(map[string]ListOfColumns, len(map[string]struct{}(tableNames)))
	for name, _ := range tableNames {
		tableInfo, err := getListOfCols(db, name)
		if err != nil {
			return nil, err
		}
		dbTablesCols[name] = tableInfo
	}

	return dbTablesCols, nil
}

func getTablesNames(db *sql.DB) (ListOfColumns, error) {
	rows, err := db.Query("SELECT table_name  FROM information_schema.tables WHERE table_schema='public' AND table_type='BASE TABLE';")
	if err != nil {
		return nil, fmt.Errorf("getTableNames : %v", err)
	}
	defer rows.Close()
	tables := make(ListOfColumns, 1)
	for rows.Next() {
		var name string
		if err = rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("getTableNames : %v", err)
		}
		tables[name] = struct{}{}
	}
	return tables, nil
}

func getListOfCols(db *sql.DB, tableName string) (ListOfColumns, error) {
	query := "SELECT * FROM " + tableName
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("getTableInfo [%s, %v]", tableName, err)
	}
	defer rows.Close()

	colNames, _ := rows.Columns()
	colsNamesSet := make(ListOfColumns, len(colNames))
	for _, n := range colNames {
		colsNamesSet[n] = struct{}{}
	}
	return colsNamesSet, nil
}
