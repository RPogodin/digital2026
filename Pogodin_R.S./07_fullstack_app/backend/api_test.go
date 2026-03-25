package backend

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestCreateTask(t *testing.T) {
	store, err := NewStore(filepath.Join(t.TempDir(), "tasks.db"))
	if err != nil {
		t.Fatalf("не удалось создать хранилище: %v", err)
	}
	defer store.Close()

	handler := NewHandler(store)

	req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewBufferString(`{
		"title":"Подготовить финальный этап",
		"description":"Собрать fullstack-демо",
		"status":"todo",
		"category":"lecture"
	}`))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("ожидался статус 201, получен %d", rec.Code)
	}
}
