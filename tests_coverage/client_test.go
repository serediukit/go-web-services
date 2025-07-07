package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// код писать тут
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

func TestWorking(t *testing.T) {
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
	if len(resp.Users) != 2 {
		t.Error("wrong result length")
	}
}
