package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
)

func NewDbExplorer(db *sql.DB) (http.Handler, error) {
	data, _ := db.Query("SELECT * FROM items")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dataToJson, _ := json.Marshal(data)
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write(dataToJson)
		fmt.Println(data)
	})

	return handler, nil
}
