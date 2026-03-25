package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateItem(t *testing.T) {
	store := newMemoryStore()
	req := httptest.NewRequest(http.MethodPost, "/api/items", bytes.NewBufferString(`{"title":"Новая задача","done":false}`))
	rec := httptest.NewRecorder()

	store.handleCreate(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("ожидался статус 201, получен %d", rec.Code)
	}
}

func TestGetUnknownItem(t *testing.T) {
	store := newMemoryStore()
	req := httptest.NewRequest(http.MethodGet, "/api/items/999", nil)
	req.SetPathValue("id", "999")
	rec := httptest.NewRecorder()

	store.handleGet(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("ожидался статус 404, получен %d", rec.Code)
	}
}
