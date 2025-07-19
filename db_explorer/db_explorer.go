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

type RowData map[string]interface{}

type ResponseItems struct {
	Tables  []string  `json:"tables,omitempty"`
	Record  RowData   `json:"record,omitempty"`
	Records []RowData `json:"records,omitempty"`
	Id      int64     `json:"id,omitempty"`
	Updated int64     `json:"updated,omitempty"`
}

type ResponseError struct {
	Error      string
	StatusCode int
}

type Resp struct {
	Response *ResponseItems `json:"response,omitempty"`
	Err      string         `json:"error,omitempty"`
}

type DBExplorer struct {
	DB *sql.DB
}

func writeResponse(w http.ResponseWriter, resp *ResponseItems, err *ResponseError) {
	w.Header().Set("Content-Type", "application/json")

	response := &Resp{}
	if err != nil {
		response.Err = err.Error
		if err.StatusCode == 0 {
			err.StatusCode = http.StatusInternalServerError
		}
		w.WriteHeader(err.StatusCode)
	} else {
		response.Response = resp
		w.WriteHeader(http.StatusOK)
	}

	resJson, _ := json.Marshal(response)

	w.Write(resJson)
}

func (dbe *DBExplorer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	url := r.URL.Path

	fmt.Printf("%7s %s\n", r.Method, r.URL.Path)

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
	} else if method == http.MethodPut {
		PutRowHandler(dbe.DB)(w, r)
	} else if method == http.MethodPost {
		PostRowHandler(dbe.DB)(w, r)
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
		res, err := GetTables(db)
		writeResponse(w, &ResponseItems{Tables: res}, err)
	}
}

func GetTables(db *sql.DB) ([]string, *ResponseError) {
	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		return nil, &ResponseError{Error: err.Error()}
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err = rows.Scan(&tableName); err != nil {
			return nil, &ResponseError{Error: err.Error()}
		}
		tables = append(tables, tableName)
	}

	return tables, nil
}

func GetRowsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		writeResponse(w, &ResponseItems{Records: res}, err)
	}
}

func GetRows(db *sql.DB, table, limit, offset string) ([]RowData, *ResponseError) {
	query := fmt.Sprintf("SELECT * FROM %s LIMIT ? OFFSET ?", table)
	rows, err := db.Query(query, limit, offset)
	if err != nil {
		return nil, &ResponseError{Error: "unknown table", StatusCode: http.StatusNotFound}
	}
	defer rows.Close()

	return unpackRows(rows)
}

func GetRowsByIDHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		urlParts := strings.Split(r.URL.Path, "/")
		table := strings.Split(urlParts[1], "?")[0]
		id := strings.Split(urlParts[2], "?")[0]

		res, err := GetRowsById(db, table, id)
		writeResponse(w, &ResponseItems{Record: res}, err)
	}
}

func GetRowsById(db *sql.DB, table, id string) (RowData, *ResponseError) {
	idColumnName, errResp := getIdColumnName(db, table)
	if errResp != nil {
		return nil, errResp
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE %s = ?", table, idColumnName)
	rows, err := db.Query(query, id)
	if err != nil {
		return nil, &ResponseError{Error: err.Error()}
	}
	defer rows.Close()

	res, errResp := unpackRows(rows)
	if errResp != nil {
		return nil, errResp
	}

	if res == nil || len(res) == 0 {
		return nil, &ResponseError{Error: "record not found", StatusCode: http.StatusNotFound}
	} else {
		return res[0], nil
	}
}

func PutRowHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		urlPart := strings.Split(r.URL.Path, "/")[1]
		table := strings.Split(urlPart, "?")[0]

		var rowData map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&rowData); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		idColumnName, errResp := getIdColumnName(db, table)
		if errResp != nil {
			writeResponse(w, nil, errResp)
			return
		}

		delete(rowData, idColumnName)

		res, err := CreateRow(db, table, rowData)
		writeResponse(w, &ResponseItems{Id: res}, err)
	}
}

func CreateRow(db *sql.DB, table string, rowData map[string]interface{}) (int64, *ResponseError) {
	query := fmt.Sprintf("INSERT INTO %s", table)

	paramRow := make([]string, len(rowData))
	valueRow := make([]string, len(rowData))

	i := 0
	for colName, val := range rowData {
		paramRow[i] = colName
		valueRow[i] = fmt.Sprintf("'%v'", val)
		i++
	}

	query += fmt.Sprintf(" (%s) VALUES (%s)", strings.Join(paramRow, ", "), strings.Join(valueRow, ", "))

	res, err := db.Exec(query)
	if err != nil {
		return -1, &ResponseError{Error: err.Error()}
	}

	id, err := res.LastInsertId()
	if err != nil {
		return -1, &ResponseError{Error: err.Error()}
	}
	return id, nil
}

func PostRowHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		urlParts := strings.Split(r.URL.Path, "/")
		table := strings.Split(urlParts[1], "?")[0]
		id := strings.Split(urlParts[2], "?")[0]

		var rowData map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&rowData); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		idColumnName, errResp := getIdColumnName(db, table)
		if errResp != nil {
			writeResponse(w, nil, errResp)
			return
		}

		if _, ok := rowData[idColumnName]; ok {
			w.WriteHeader(http.StatusBadRequest)
			writeResponse(w, nil, &ResponseError{Error: fmt.Sprintf("field %s have invalid type", idColumnName), StatusCode: http.StatusBadRequest})
			return
		}

		res, err := UpdateRow(db, table, id, idColumnName, rowData)
		writeResponse(w, &ResponseItems{Updated: res}, err)
	}
}

func UpdateRow(db *sql.DB, table, id, idColumnName string, rowData map[string]interface{}) (int64, *ResponseError) {
	updateRow := make([]string, len(rowData))

	i := 0
	for colName, val := range rowData {
		updateRow[i] = fmt.Sprintf("%s = '%v'", colName, val)
		i++
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s = %s", table, strings.Join(updateRow, ", "), idColumnName, id)

	_, err := db.Exec(query)
	if err != nil {
		return -1, &ResponseError{Error: err.Error()}
	}
	return 1, nil
}

func unpackRows(rows *sql.Rows) ([]RowData, *ResponseError) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, &ResponseError{Error: err.Error()}
	}

	res := make([]RowData, 0)

	for rows.Next() {
		row := NewRow(len(columns))

		for i := range columns {
			row.ColumnPointers[i] = &row.ColumnValues[i]
		}

		if err = rows.Scan(row.ColumnPointers...); err != nil {
			return nil, &ResponseError{Error: err.Error()}
		}

		rowData := make(map[string]interface{})
		for i, colName := range columns {
			val := row.ColumnValues[i]
			if b, ok := val.([]byte); ok {
				if string(b) == "<nil>" {
					rowData[colName] = nil
				} else {
					rowData[colName] = string(b)
				}
			} else {
				rowData[colName] = val
			}
		}

		res = append(res, rowData)
	}

	return res, nil
}

func getIdColumnName(db *sql.DB, table string) (string, *ResponseError) {
	rows, err := db.Query("SELECT COLUMN_NAME FROM information_schema.KEY_COLUMN_USAGE WHERE TABLE_NAME = ? AND CONSTRAINT_NAME = 'PRIMARY'", table)
	if err != nil {
		return "", &ResponseError{Error: err.Error()}
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			return "", &ResponseError{Error: err.Error()}
		}
		return id, nil
	}

	return "", nil
}
