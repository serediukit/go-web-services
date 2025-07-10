package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

func (obj *MyApi) Unpack(params url.Values) error {

	// nextID
	nextIDRaw, err := strconv.ParseUint(params.Get("nextID"), 10, 64)
	if err != nil {
		return ApiError{http.StatusBadRequest, fmt.Errorf("invalid nextID - must be uint64")}
	}
	obj.nextID = nextIDRaw

	return nil
}

func (obj *ProfileParams) Unpack(params url.Values) error {

	// Login
	LoginRaw := params.Get("Login")
	obj.Login = LoginRaw

	return nil
}

func (obj *CreateParams) Unpack(params url.Values) error {

	// Login
	LoginRaw := params.Get("Login")
	obj.Login = LoginRaw

	// Name
	NameRaw := params.Get("Name")
	obj.Name = NameRaw

	// Status
	StatusRaw := params.Get("Status")
	obj.Status = StatusRaw

	// Age
	AgeRaw, err := strconv.Atoi(params.Get("Age"))
	if err != nil {
		return ApiError{http.StatusBadRequest, fmt.Errorf("invalid Age - must be int")}
	}
	obj.Age = AgeRaw

	return nil
}

func (obj *User) Unpack(params url.Values) error {

	// ID
	IDRaw, err := strconv.ParseUint(params.Get("ID"), 10, 64)
	if err != nil {
		return ApiError{http.StatusBadRequest, fmt.Errorf("invalid ID - must be uint64")}
	}
	obj.ID = IDRaw

	// Login
	LoginRaw := params.Get("Login")
	obj.Login = LoginRaw

	// FullName
	FullNameRaw := params.Get("FullName")
	obj.FullName = FullNameRaw

	// Status
	StatusRaw, err := strconv.Atoi(params.Get("Status"))
	if err != nil {
		return ApiError{http.StatusBadRequest, fmt.Errorf("invalid Status - must be int")}
	}
	obj.Status = StatusRaw

	return nil
}

func (obj *NewUser) Unpack(params url.Values) error {

	// ID
	IDRaw, err := strconv.ParseUint(params.Get("ID"), 10, 64)
	if err != nil {
		return ApiError{http.StatusBadRequest, fmt.Errorf("invalid ID - must be uint64")}
	}
	obj.ID = IDRaw

	return nil
}

func (obj *OtherApi) Unpack(params url.Values) error {

	return nil
}

func (obj *OtherCreateParams) Unpack(params url.Values) error {

	// Username
	UsernameRaw := params.Get("Username")
	obj.Username = UsernameRaw

	// Name
	NameRaw := params.Get("Name")
	obj.Name = NameRaw

	// Class
	ClassRaw := params.Get("Class")
	obj.Class = ClassRaw

	// Level
	LevelRaw, err := strconv.Atoi(params.Get("Level"))
	if err != nil {
		return ApiError{http.StatusBadRequest, fmt.Errorf("invalid Level - must be int")}
	}
	obj.Level = LevelRaw

	return nil
}

func (obj *OtherUser) Unpack(params url.Values) error {

	// ID
	IDRaw, err := strconv.ParseUint(params.Get("ID"), 10, 64)
	if err != nil {
		return ApiError{http.StatusBadRequest, fmt.Errorf("invalid ID - must be uint64")}
	}
	obj.ID = IDRaw

	// Login
	LoginRaw := params.Get("Login")
	obj.Login = LoginRaw

	// FullName
	FullNameRaw := params.Get("FullName")
	obj.FullName = FullNameRaw

	// Level
	LevelRaw, err := strconv.Atoi(params.Get("Level"))
	if err != nil {
		return ApiError{http.StatusBadRequest, fmt.Errorf("invalid Level - must be int")}
	}
	obj.Level = LevelRaw

	return nil
}
