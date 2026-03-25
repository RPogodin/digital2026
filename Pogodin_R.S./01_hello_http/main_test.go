package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHelloHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	helloHandler(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("ожидался статус 200, получен %d", res.StatusCode)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Hello, HTTP!") {
		t.Fatalf("ожидался текст приветствия, получено %q", body)
	}
}
