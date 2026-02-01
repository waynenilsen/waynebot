package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPFetchSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("ok"))
	}))
	defer srv.Close()

	fn := HTTPFetch()

	args, _ := json.Marshal(httpFetchArgs{URL: srv.URL})
	out, err := fn(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	if out != "ok" {
		t.Fatalf("got %q, want %q", out, "ok")
	}
}

func TestHTTPFetchEmptyURL(t *testing.T) {
	fn := HTTPFetch()

	args, _ := json.Marshal(httpFetchArgs{})
	_, err := fn(context.Background(), args)
	if err == nil {
		t.Fatal("expected error for empty url")
	}
}

func TestHTTPFetchResponseCap(t *testing.T) {
	big := strings.Repeat("x", httpFetchMaxResponse+1000)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(big))
	}))
	defer srv.Close()

	fn := HTTPFetch()

	args, _ := json.Marshal(httpFetchArgs{URL: srv.URL})
	out, err := fn(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "truncated") {
		t.Fatal("expected truncation notice")
	}
}

func TestHTTPFetchHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
	}))
	defer srv.Close()

	fn := HTTPFetch()

	args, _ := json.Marshal(httpFetchArgs{URL: srv.URL})
	out, err := fn(context.Background(), args)
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "not found" {
		t.Fatalf("body = %q", out)
	}
}
