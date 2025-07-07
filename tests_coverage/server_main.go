package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"
)

type XmlRow struct {
	Id        int    `xml:"id"`
	FirstName string `xml:"first_name"`
	LastName  string `xml:"last_name"`
	Age       int    `xml:"age"`
	About     string `xml:"about"`
	Gender    string `xml:"gender"`
}

type XmlData struct {
	Users []User
}

const AccessToken = "access_token_123_456"

// main() is used for manual server listening
//func main() {
//	http.HandleFunc("/", SearchServer)
//
//	fmt.Println("Starting server on port 8080")
//	err := http.ListenAndServe(":8080", nil)
//	if err != nil {
//		return
//	}
//}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if r.Header.Get("AccessToken") != AccessToken {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	data := &XmlData{}

	query := r.URL.Query()

	err := data.load(query.Get("query"))
	if err != nil {
		sendError(w, http.StatusBadRequest, err)
		return
	}

	orderField := query.Get("order_field")
	orderBy, _ := strconv.Atoi(query.Get("order_by"))
	err = data.sort(orderField, orderBy)
	if err != nil {
		sendError(w, http.StatusBadRequest, err)
		return
	}

	limit, _ := strconv.Atoi(query.Get("limit"))
	offset, _ := strconv.Atoi(query.Get("offset"))
	err = data.setLimitOffset(limit, offset)
	if err != nil {
		sendError(w, http.StatusBadRequest, err)
		return
	}

	result, err := json.Marshal(data.Users)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(result)
}

func sendError(w http.ResponseWriter, code int, err error) {
	data, err := json.Marshal(SearchErrorResponse{Error: err.Error()})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(data)
}

func (data *XmlData) load(query string) error {
	fileData, err := os.ReadFile("dataset.xml")
	if err != nil {
		return err
	}

	decoder := xml.NewDecoder(bytes.NewReader(fileData))

	for {
		t, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if start, ok := t.(xml.StartElement); ok && start.Name.Local == "row" {
			var row XmlRow
			err := decoder.DecodeElement(&row, &start)
			if err != nil {
				return err
			}

			user := User{
				Id:     row.Id,
				Name:   row.FirstName + " " + row.LastName,
				Age:    row.Age,
				About:  row.About,
				Gender: row.Gender,
			}

			if strings.Contains(user.Name, query) || strings.Contains(row.About, query) {
				data.Users = append(data.Users, user)
			}
		}
	}
}

func (data *XmlData) sort(orderField string, orderBy int) error {
	if !slices.Contains([]int{OrderByAsc, OrderByAsIs, OrderByDesc}, orderBy) {
		return fmt.Errorf("OrderBy invalid")
	}
	if !slices.Contains([]string{"Id", "Name", "Age", ""}, orderField) {
		return fmt.Errorf(ErrorBadOrderField)
	}
	if orderBy != OrderByAsIs {
		sort.Slice(
			data.Users,
			func(i, j int) bool {
				var res bool
				switch orderField {
				case "Id":
					res = data.Users[i].Id < data.Users[j].Id
				case "Age":
					res = data.Users[i].Age < data.Users[j].Age
				case "":
					fallthrough
				case "Name":
					res = data.Users[i].Name < data.Users[j].Name
				}
				if orderBy == OrderByAsc {
					return res
				} else {
					return !res
				}
			},
		)
	}
	return nil
}

func (data *XmlData) setLimitOffset(limit int, offset int) error {
	if limit > 0 {
		if offset > len(data.Users)-1 {
			data.Users = []User{}
		} else if offset+limit > len(data.Users) {
			data.Users = data.Users[offset:]
		} else {
			data.Users = data.Users[offset : offset+limit]
		}
		return nil
	}
	return fmt.Errorf("limit invalid")
}

func (data *XmlData) Print() {
	encoder := xml.NewEncoder(os.Stdout)
	encoder.Indent("", "  ")
	err := encoder.Encode(data)
	if err != nil {
		panic(err)
	}
}
