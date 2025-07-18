package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Row struct {
	ColumnPointers []interface{}
	ColumnValues   []interface{}
}

func NewRow(l int) Row {
	return Row{
		ColumnPointers: make([]interface{}, l),
		ColumnValues:   make([]interface{}, l),
	}
}

type RowsData []map[string]interface{}

type DBExplorer struct {
	DB *sql.DB
}

func (dbe *DBExplorer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	url := r.URL.Path
	fmt.Println(url)

	if method == http.MethodGet {
		urlParts := strings.Split(url, "/")

		switch len(urlParts) {
		case 2:
			if len(urlParts[1]) == 0 {
				GetTablesHandler(dbe.DB)(w, r)
			} else {
				GetRowsHandler(dbe.DB)(w, r)
			}
		case 3:
			GetRowsByIDHandler(dbe.DB)(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	} else if method == http.MethodPost {

	} else if method == http.MethodPut {

	} else if method == http.MethodDelete {

	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func NewDbExplorer(db *sql.DB) (http.Handler, error) {
	return &DBExplorer{DB: db}, nil
}

func GetTablesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("GET TABLES")

		res, err := GetTables(db)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		resJson, err := json.Marshal(res)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(resJson)
	}
}

func GetTables(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables := []string{}
	for rows.Next() {
		var tableName string
		if err = rows.Scan(&tableName); err != nil {
			return nil, err
		}
		tables = append(tables, tableName)
	}

	return tables, nil
}

func GetRowsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("GET ROWS")

		urlPart := strings.Split(r.URL.Path, "/")[1]
		table := strings.Split(urlPart, "?")[0]

		limit := r.URL.Query().Get("limit")
		if limit == "" {
			limit = "5"
		}

		offset := r.URL.Query().Get("offset")
		if offset == "" {
			offset = "0"
		}

		res, err := GetRows(db, table, limit, offset)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		fmt.Println(res)

		resJson, err := json.Marshal(res)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(resJson)
	}
}

func GetRows(db *sql.DB, table, limit, offset string) (RowsData, error) {
	query := fmt.Sprintf("SELECT * FROM %s LIMIT ? OFFSET ?", table)
	rows, err := db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return unpackRows(rows)
}

func GetRowsByIDHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("GET ROWS BY ID")

		urlParts := strings.Split(r.URL.Path, "/")
		table := strings.Split(urlParts[1], "?")[0]
		id := strings.Split(urlParts[2], "?")[0]

		res, err := GetRowsById(db, table, id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		fmt.Println(res)

		resJson, err := json.Marshal(res)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(resJson)
	}
}

func GetRowsById(db *sql.DB, table, id string) (RowsData, error) {
	idColumnName, err := getIdColumnName(db, table)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE %s = ?", table, idColumnName)
	rows, err := db.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return unpackRows(rows)
}

func unpackRows(rows *sql.Rows) (RowsData, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	res := make(RowsData, 0)

	for rows.Next() {
		row := NewRow(len(columns))

		for i := range columns {
			row.ColumnPointers[i] = &row.ColumnValues[i]
		}

		if err = rows.Scan(row.ColumnPointers...); err != nil {
			return nil, err
		}

		rowData := make(map[string]interface{})
		for i, colName := range columns {
			val := row.ColumnValues[i]
			if b, ok := val.([]byte); ok {
				rowData[colName] = string(b)
			} else {
				rowData[colName] = val
			}
		}

		res = append(res, rowData)
	}

	return res, nil
}

func getIdColumnName(db *sql.DB, table string) (string, error) {
	rows, err := db.Query("SELECT COLUMN_NAME FROM information_schema.KEY_COLUMN_USAGE WHERE TABLE_NAME = ? AND CONSTRAINT_NAME = 'PRIMARY'", table)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			return "", err
		}
		return id, nil
	}

	return "", nil
}
