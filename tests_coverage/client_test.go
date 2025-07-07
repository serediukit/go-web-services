package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
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

func TestCorrect(t *testing.T) {
	ts := NewTestServer(AccessToken)
	defer ts.Close()

	resp, err := ts.client.FindUsers(goodRequest)
	if err != nil {
		t.Error(err)
	}
	if resp == nil {
		t.Error("response is nil")
	}
	if len(resp.Users) != 2 {
		t.Errorf("response length mismatch: expected 2, got %d", len(resp.Users))
	}
	if !resp.NextPage {
		t.Error("next page is false")
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

func TestIncorrectToken(t *testing.T) {
	ts := NewTestServer("incorrect_token")
	defer ts.Close()

	_, err := ts.client.FindUsers(goodRequest)
	if err != nil && err.Error() != "Bad AccessToken" {
		t.Error(err)
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
			err:  fmt.Errorf("unknown bad request error: %s", ErrorBadOrderField),
		},
		{
			name: "bad order field register",
			req:  SearchRequest{OrderField: "name"},
			err:  fmt.Errorf("unknown bad request error: %s", ErrorBadOrderField),
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
