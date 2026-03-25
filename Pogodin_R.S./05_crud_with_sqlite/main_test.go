package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestCreateAndListItems(t *testing.T) {
	db, err := openDB(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("не удалось открыть тестовую БД: %v", err)
	}
	defer db.Close()

	server := &app{db: db}

	createReq := httptest.NewRequest(http.MethodPost, "/api/items", bytes.NewBufferString(`{"title":"Проверить SQLite","done":false}`))
	createRec := httptest.NewRecorder()
	server.handleCreate(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("ожидался статус 201, получен %d", createRec.Code)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/items", nil)
	listRec := httptest.NewRecorder()
	server.handleList(listRec, listReq)

	if listRec.Code != http.StatusOK {
		t.Fatalf("ожидался статус 200, получен %d", listRec.Code)
	}
}
