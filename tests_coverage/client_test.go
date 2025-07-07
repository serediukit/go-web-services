package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

var goodRequest = SearchRequest{
	Query:      "ipsum",
	OrderField: "Id",
	OrderBy:    OrderByAsc,
	Limit:      2,
	Offset:     1,
}

type TestCase struct {
	name        string
	accessToken string
	req         SearchRequest
	resp        *SearchResponse
	err         error
}

type TestServer struct {
	client *SearchClient
	server *httptest.Server
}

func (srv *TestServer) Close() {
	srv.server.Close()
}

func NewTestServer(token string) *TestServer {
	server := httptest.NewServer(http.HandlerFunc(SearchServer))
	return &TestServer{
		&SearchClient{token, server.URL},
		server,
	}
}

func TestWithNextPage(t *testing.T) {
	ts := NewTestServer(AccessToken)
	defer ts.Close()

	resp, err := ts.client.FindUsers(goodRequest)
	if err != nil {
		t.Error(err)
	} else if resp == nil {
		t.Error("response is nil")
	} else if len(resp.Users) != 2 {
		t.Errorf("response length mismatch: expected 2, got %d", len(resp.Users))
	} else if !resp.NextPage {
		t.Error("next page is false")
	}
}

func TestWithoutNextPage(t *testing.T) {
	ts := NewTestServer(AccessToken)
	defer ts.Close()

	req := goodRequest
	req.Limit = 100

	resp, err := ts.client.FindUsers(req)
	if err != nil {
		t.Error(err)
	}
	if resp == nil {
		t.Error("response is nil")
	} else if len(resp.Users) != 23 {
		t.Errorf("response length mismatch: expected 23, got %d", len(resp.Users))
	} else if resp.NextPage {
		t.Error("next page is true")
	}
}

func TestOrderField(t *testing.T) {
	orderFields := []string{"Id", "Name", "Age", ""}

	ts := NewTestServer(AccessToken)
	defer ts.Close()

	for _, orderField := range orderFields {
		req := goodRequest
		req.OrderField = orderField
		resp, err := ts.client.FindUsers(req)
		if err != nil {
			t.Error(err)
		}
		if resp == nil {
			t.Error("response is nil")
		}
	}
}

func TestInvalidServer(t *testing.T) {
	ts := NewTestServer(AccessToken)
	defer ts.Close()

	ts.client.URL = "http://invalid.server.com"

	_, err := ts.client.FindUsers(goodRequest)
	if err == nil {
		t.Error("error is nil")
	} else if !strings.Contains(err.Error(), "unknown error") {
		t.Errorf("Invalid error: %v", err.Error())
	}
}

func TestTimeout(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
	}))
	defer s.Close()

	c := &SearchClient{AccessToken, s.URL}

	_, err := c.FindUsers(goodRequest)
	if err == nil {
		t.Error("error is nil")
	} else if !strings.Contains(err.Error(), "timeout for") {
		t.Errorf("Invalid error: %v", err.Error())
	}
}

func TestInvalidToken(t *testing.T) {
	ts := NewTestServer("incorrect_token")
	defer ts.Close()

	_, err := ts.client.FindUsers(goodRequest)
	if err == nil {
		t.Error("error is nil")
	} else if err.Error() != "Bad AccessToken" {
		t.Errorf("Invalid error: %v", err.Error())
	}
}

func TestCantUnpack(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Some Error", http.StatusBadRequest)
	}))
	c := SearchClient{AccessToken, s.URL}
	defer s.Close()

	_, err := c.FindUsers(SearchRequest{})

	if err == nil {
		t.Errorf("error is nil")
	} else if !strings.Contains(err.Error(), "cant unpack error json") {
		t.Errorf("Invalid error: %v", err.Error())
	}
}

func TestCantUnpackResult(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("None"))
	}))
	defer s.Close()

	c := &SearchClient{AccessToken, s.URL}

	_, err := c.FindUsers(goodRequest)
	if err == nil {
		t.Error("error is nil")
	} else if !strings.Contains(err.Error(), "cant unpack result json") {
		t.Errorf("Invalid error: %v", err.Error())
	}
}

func TestUnknownBadRequestError(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sendError(w, http.StatusBadRequest, fmt.Errorf("bad request"))
	}))
	c := SearchClient{AccessToken, s.URL}
	defer s.Close()

	_, err := c.FindUsers(SearchRequest{})

	if err == nil {
		t.Errorf("error is nil")
	} else if !strings.Contains(err.Error(), "unknown bad request error") {
		t.Errorf("Invalid error: %v", err.Error())
	}
}

func TestFatalError(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Fatal Error", http.StatusInternalServerError)
	}))
	c := SearchClient{AccessToken, s.URL}
	defer s.Close()

	_, err := c.FindUsers(SearchRequest{})

	if err == nil {
		t.Errorf("Empty error")
	} else if err.Error() != "SearchServer fatal error" {
		t.Errorf("Invalid error: %v", err.Error())
	}
}

func TestErrorReceiving(t *testing.T) {
	testCases := []TestCase{
		{
			name: "limit under 0",
			req:  SearchRequest{Limit: -1},
			err:  fmt.Errorf("limit must be > 0"),
		},
		{
			name: "limit upper 25",
			req:  SearchRequest{Limit: 100},
		},
		{
			name: "offset under 0",
			req:  SearchRequest{Limit: 5, Offset: -1},
			err:  fmt.Errorf("offset must be > 0"),
		},
		{
			name: "bad order field",
			req:  SearchRequest{OrderField: "bad_field"},
			err:  fmt.Errorf("OrderFeld bad_field invalid"),
		},
		{
			name: "bad order field register",
			req:  SearchRequest{OrderField: "name"},
			err:  fmt.Errorf("OrderFeld name invalid"),
		},
	}

	ts := NewTestServer(AccessToken)
	defer ts.Close()

	for _, testCase := range testCases {
		_, err := ts.client.FindUsers(testCase.req)
		if err != nil && err.Error() != testCase.err.Error() {
			t.Errorf("error in test - %s", testCase.name)
			t.Errorf("Request: %v", testCase.req)
			t.Errorf("Response: %v", testCase.resp)
			t.Errorf("error mismatch: expected \"%s\", got \"%s\"", testCase.err, err)
		}
	}
}
