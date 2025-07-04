package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"sort"
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

// код писать тут
func main() {
	params := SearchRequest{
		Limit:      10,
		Offset:     0,
		Query:      "ipsum",
		OrderField: "Id",
		OrderBy:    OrderByAsc,
	}

	data := readXml(params.Query)

	if params.OrderBy != OrderByAsIs {
		sort.Slice(
			data.Users,
			func(i, j int) bool {
				var res bool
				switch params.OrderField {
				case "Id":
					res = data.Users[i].Id < data.Users[j].Id
				case "Name":
					fallthrough
				case "":
					res = data.Users[i].Name < data.Users[j].Name
				case "Age":
					res = data.Users[i].Age < data.Users[j].Age
				default:
					panic(fmt.Errorf(ErrorBadOrderField))
				}
				if params.OrderBy == OrderByAsc {
					return res
				} else {
					return !res
				}
			},
		)
	}

	if params.Limit > 0 {
		if params.Offset > len(data.Users)-1 {
			data.Users = []User{}
		} else if params.Offset+params.Limit > len(data.Users) {
			data.Users = data.Users[params.Offset:]
		} else {
			data.Users = data.Users[params.Offset : params.Offset+params.Limit]
		}
	}
}

func readXml(query string) *XmlData {
	fileData, err := os.ReadFile("dataset.xml")
	if err != nil {
		panic(err)
	}

	decoder := xml.NewDecoder(bytes.NewReader(fileData))
	var data XmlData

	for {
		t, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		if start, ok := t.(xml.StartElement); ok && start.Name.Local == "row" {
			var row XmlRow
			err := decoder.DecodeElement(&row, &start)
			if err != nil {
				panic(err)
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

	return &data
}
