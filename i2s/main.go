package main

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type Simple struct {
	ID       int
	Username string
	Active   bool
}

type IDBlock struct {
	ID int
}

func main() {
	expected := &Simple{
		ID:       42,
		Username: "rvasily",
		Active:   true,
	}
	jsonRaw, _ := json.Marshal(expected)
	// fmt.Println(string(jsonRaw))

	var tmpData interface{}
	json.Unmarshal(jsonRaw, &tmpData)

	result := new(Simple)
	err := i2s(tmpData, result)

	if err != nil {
		fmt.Println(fmt.Errorf("unexpected error: %v", err))
	}
	if !reflect.DeepEqual(expected, result) {
		fmt.Println(fmt.Errorf("results not match\nGot:\n%#v\nExpected:\n%#v", result, expected))
	}
}
