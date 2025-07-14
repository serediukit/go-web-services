package main

import (
	"encoding/json"
	"errors"
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
		return ApiError{http.StatusBadRequest, fmt.Errorf("login must be not empty")}
	}

	return nil
}

func (obj *CreateParams) Unpack(params url.Values) error {

	// Login
	LoginRaw := params.Get("login")
	obj.Login = LoginRaw

	// Name
	NameRaw := params.Get("full_name")
	obj.Name = NameRaw

	// Status
	StatusRaw := params.Get("status")
	obj.Status = StatusRaw

	// Age
	AgeRaw, err := strconv.Atoi(params.Get("age"))
	if err != nil {
		return ApiError{http.StatusBadRequest, fmt.Errorf("age must be int")}
	}
	obj.Age = AgeRaw

	return nil
}

func (obj *CreateParams) Validate() error {

	// Age min
	if obj.Age < 0 {
		return ApiError{http.StatusBadRequest, fmt.Errorf("age must be >= 0")}
	}

	// Age max
	if obj.Age > 128 {
		return ApiError{http.StatusBadRequest, fmt.Errorf("age must be <= 128")}
	}

	// Login required
	if obj.Login == "" {
		return ApiError{http.StatusBadRequest, fmt.Errorf("login must be not empty")}
	}

	// Login min
	if len(obj.Login) < 10 {
		return ApiError{http.StatusBadRequest, fmt.Errorf("login len must be >= 10")}
	}

	// Status default
	if obj.Status == "" {
		obj.Status = "user"
	}

	// Status enum
	if !slices.Contains([]string{"user", "moderator", "admin"}, obj.Status) {
		return ApiError{http.StatusBadRequest, fmt.Errorf("status must be one of [user, moderator, admin]")}
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
		return nil, ApiError{http.StatusForbidden, fmt.Errorf("unauthorized")}
	}

	if r.Method != "POST" {
		return nil, ApiError{http.StatusNotAcceptable, fmt.Errorf("bad method")}
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

	// Name
	NameRaw := params.Get("account_name")
	obj.Name = NameRaw

	// Class
	ClassRaw := params.Get("class")
	obj.Class = ClassRaw

	// Level
	LevelRaw, err := strconv.Atoi(params.Get("level"))
	if err != nil {
		return ApiError{http.StatusBadRequest, fmt.Errorf("level must be int")}
	}
	obj.Level = LevelRaw

	return nil
}

func (obj *OtherCreateParams) Validate() error {

	// Class default
	if obj.Class == "" {
		obj.Class = "warrior"
	}

	// Class enum
	if !slices.Contains([]string{"warrior", "sorcerer", "rouge"}, obj.Class) {
		return ApiError{http.StatusBadRequest, fmt.Errorf("class must be one of [warrior, sorcerer, rouge]")}
	}

	// Level min
	if obj.Level < 1 {
		return ApiError{http.StatusBadRequest, fmt.Errorf("level must be >= 1")}
	}

	// Level max
	if obj.Level > 50 {
		return ApiError{http.StatusBadRequest, fmt.Errorf("level must be <= 50")}
	}

	// Username required
	if obj.Username == "" {
		return ApiError{http.StatusBadRequest, fmt.Errorf("username must be not empty")}
	}

	// Username min
	if len(obj.Username) < 3 {
		return ApiError{http.StatusBadRequest, fmt.Errorf("username len must be >= 3")}
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
		return nil, ApiError{http.StatusForbidden, fmt.Errorf("unauthorized")}
	}

	if r.Method != "POST" {
		return nil, ApiError{http.StatusNotAcceptable, fmt.Errorf("bad method")}
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

func (h *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		err error
		res interface{}
	)

	switch r.URL.Path {
	case "/user/profile":
		res, err = h.wrapperProfile(w, r)
	case "/user/create":
		res, err = h.wrapperCreate(w, r)
	default:
		err = ApiError{http.StatusNotFound, fmt.Errorf("unknown method")}
	}

	var response = struct {
		Error    string      `json:"error"`
		Response interface{} `json:"response,omitempty"`
	}{}

	if err == nil {
		response.Response = res
	} else {
		response.Error = err.Error()

		var errApi ApiError
		if errors.As(err, &errApi) {
			w.WriteHeader(errApi.HTTPStatus)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	responseJson, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJson)
}

func (h *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		err error
		res interface{}
	)

	switch r.URL.Path {
	case "/user/create":
		res, err = h.wrapperCreate(w, r)
	default:
		err = ApiError{http.StatusNotFound, fmt.Errorf("unknown method")}
	}

	var response = struct {
		Error    string      `json:"error"`
		Response interface{} `json:"response,omitempty"`
	}{}

	if err == nil {
		response.Response = res
	} else {
		response.Error = err.Error()

		var errApi ApiError
		if errors.As(err, &errApi) {
			w.WriteHeader(errApi.HTTPStatus)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	responseJson, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJson)
}
