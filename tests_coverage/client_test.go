package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

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

func TestReceiving(t *testing.T) {
	ts := NewTestServer(AccessToken)
	defer ts.Close()

	resp, err := ts.client.FindUsers(SearchRequest{
		Query:      "ipsum",
		OrderField: "Id",
		OrderBy:    OrderByAsc,
		Limit:      2,
		Offset:     1,
	})
	if err != nil {
		t.Error(err)
	}
	if resp == nil {
		t.Error("response is nil")
	}
}

func TestErrorReceiving(t *testing.T) {
	testCases := []TestCase{
		{
			name: "limit under 0",
			req:  SearchRequest{Limit: -1},
			err:  fmt.Errorf("limit must be > 0"),
		},
	}

	for _, testCase := range testCases {
		at := AccessToken
		if testCase.accessToken != "" {
			at = testCase.accessToken
		}

		ts := NewTestServer(at)
		defer ts.Close()

		_, err := ts.client.FindUsers(testCase.req)
		if err.Error() != testCase.err.Error() {
			t.Errorf("error in test - %s", testCase.name)
			t.Errorf("Request: %v", testCase.req)
			t.Errorf("Response: %v", testCase.resp)
			t.Errorf("error mismatch: expected \"%s\", got \"%s\"", testCase.err, err)
		}
	}
}
