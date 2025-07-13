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

	// Login required
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

	// Age max
	if obj.Age > 128 {
		return ApiError{http.StatusBadRequest, fmt.Errorf("invalid Age - must be less than max")}
	}

	// Login required
	if obj.Login == "" {
		return ApiError{http.StatusBadRequest, fmt.Errorf("invalid Login - field is required")}
	}

	// Login min
	if len(obj.Login) < 10 {
		return ApiError{http.StatusBadRequest, fmt.Errorf("invalid Login - len must be more than min")}
	}

	// Status enum
	if !slices.Contains([]string{"user", "moderator", "admin"}, obj.Status) {
		return ApiError{http.StatusBadRequest, fmt.Errorf("invalid Status - must be in enum")}
	}

	// Status default
	if obj.Status == "" {
		obj.Status = "user"
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

func (h *MyApi) wrapperProfile(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	var params url.Values
	if r.Method == "GET" {
		params = r.URL.Query()
	} else {
		err := r.ParseForm()
		if err != nil {
			return nil, ApiError{http.StatusBadRequest, fmt.Errorf("invalid request")}
		}
		params = r.PostForm
	}

	in := ProfileParams{}
	err := in.Unpack(params)
	if err != nil {
		return nil, ApiError{http.StatusBadRequest, err}
	}

	err = in.Validate()
	if err != nil {
		return nil, ApiError{http.StatusBadRequest, err}
	}

	return h.Profile(r.Context(), in)
}

func (h *MyApi) wrapperCreate(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	if r.Header.Get("X-Auth") != "100500" {
		return nil, ApiError{http.StatusUnauthorized, fmt.Errorf("unauthorized")}
	}

	if r.Method != "POST" {
		return nil, ApiError{http.StatusMethodNotAllowed, fmt.Errorf("method not allowed")}
	}

	var params url.Values
	if r.Method == "GET" {
		params = r.URL.Query()
	} else {
		err := r.ParseForm()
		if err != nil {
			return nil, ApiError{http.StatusBadRequest, fmt.Errorf("invalid request")}
		}
		params = r.PostForm
	}

	in := CreateParams{}
	err := in.Unpack(params)
	if err != nil {
		return nil, ApiError{http.StatusBadRequest, err}
	}

	err = in.Validate()
	if err != nil {
		return nil, ApiError{http.StatusBadRequest, err}
	}

	return h.Create(r.Context(), in)
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

	// Username required
	if obj.Username == "" {
		return ApiError{http.StatusBadRequest, fmt.Errorf("invalid Username - field is required")}
	}

	// Username min
	if len(obj.Username) < 3 {
		return ApiError{http.StatusBadRequest, fmt.Errorf("invalid Username - len must be more than min")}
	}

	// Class enum
	if !slices.Contains([]string{"warrior", "sorcerer", "rouge"}, obj.Class) {
		return ApiError{http.StatusBadRequest, fmt.Errorf("invalid Class - must be in enum")}
	}

	// Class default
	if obj.Class == "" {
		obj.Class = "warrior"
	}

	// Level min
	if obj.Level < 1 {
		return ApiError{http.StatusBadRequest, fmt.Errorf("invalid Level - must be more than min")}
	}

	// Level max
	if obj.Level > 50 {
		return ApiError{http.StatusBadRequest, fmt.Errorf("invalid Level - must be less than max")}
	}

	return nil
}

func (obj *OtherUser) Unpack(params url.Values) error {

	return nil
}

func (obj *OtherUser) Validate() error {

	return nil
}

func (h *OtherApi) wrapperCreate(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	if r.Header.Get("X-Auth") != "100500" {
		return nil, ApiError{http.StatusUnauthorized, fmt.Errorf("unauthorized")}
	}

	if r.Method != "POST" {
		return nil, ApiError{http.StatusMethodNotAllowed, fmt.Errorf("method not allowed")}
	}

	var params url.Values
	if r.Method == "GET" {
		params = r.URL.Query()
	} else {
		err := r.ParseForm()
		if err != nil {
			return nil, ApiError{http.StatusBadRequest, fmt.Errorf("invalid request")}
		}
		params = r.PostForm
	}

	in := OtherCreateParams{}
	err := in.Unpack(params)
	if err != nil {
		return nil, ApiError{http.StatusBadRequest, err}
	}

	err = in.Validate()
	if err != nil {
		return nil, ApiError{http.StatusBadRequest, err}
	}

	return h.Create(r.Context(), in)
}
