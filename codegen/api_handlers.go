package main

import (
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strconv"
)

func (obj *MyApi) Unpack(params url.Values) error {

	return nil
}

func (obj *MyApi) Validate() error {

	return nil
}

func (obj *ProfileParams) Unpack(params url.Values) error {

	// Login
	LoginRaw := params.Get("login")
	obj.Login = LoginRaw

	return nil
}

func (obj *ProfileParams) Validate() error {

	// required
	if obj.Login == "" {
		return ApiError{http.StatusBadRequest, fmt.Errorf("invalid Login - field is required")}
	}

	return nil
}

func (obj *CreateParams) Unpack(params url.Values) error {

	// Login
	LoginRaw := params.Get("login")
	obj.Login = LoginRaw

	// Status
	StatusRaw := params.Get("status")
	obj.Status = StatusRaw

	// Age
	AgeRaw, err := strconv.Atoi(params.Get("age"))
	if err != nil {
		return ApiError{http.StatusBadRequest, fmt.Errorf("invalid Age - must be int")}
	}
	obj.Age = AgeRaw

	return nil
}

func (obj *CreateParams) Validate() error {

	// required
	if obj.Login == "" {
		return ApiError{http.StatusBadRequest, fmt.Errorf("invalid Login - field is required")}
	}

	// enum
	if !slices.Contains([]string{"user", "moderator", "admin"}, obj.Status) {
		return ApiError{http.StatusBadRequest, fmt.Errorf("invalid Status - must be in enum")}
	}

	return nil
}

func (obj *User) Unpack(params url.Values) error {

	return nil
}

func (obj *User) Validate() error {

	return nil
}

func (obj *NewUser) Unpack(params url.Values) error {

	return nil
}

func (obj *NewUser) Validate() error {

	return nil
}

func (obj *OtherApi) Unpack(params url.Values) error {

	return nil
}

func (obj *OtherApi) Validate() error {

	return nil
}

func (obj *OtherCreateParams) Unpack(params url.Values) error {

	// Username
	UsernameRaw := params.Get("username")
	obj.Username = UsernameRaw

	// Class
	ClassRaw := params.Get("class")
	obj.Class = ClassRaw

	// Level
	LevelRaw, err := strconv.Atoi(params.Get("level"))
	if err != nil {
		return ApiError{http.StatusBadRequest, fmt.Errorf("invalid Level - must be int")}
	}
	obj.Level = LevelRaw

	return nil
}

func (obj *OtherCreateParams) Validate() error {

	// required
	if obj.Username == "" {
		return ApiError{http.StatusBadRequest, fmt.Errorf("invalid Username - field is required")}
	}

	// enum
	if !slices.Contains([]string{"warrior", "sorcerer", "rouge"}, obj.Class) {
		return ApiError{http.StatusBadRequest, fmt.Errorf("invalid Class - must be in enum")}
	}

	return nil
}

func (obj *OtherUser) Unpack(params url.Values) error {

	return nil
}

func (obj *OtherUser) Validate() error {

	return nil
}
