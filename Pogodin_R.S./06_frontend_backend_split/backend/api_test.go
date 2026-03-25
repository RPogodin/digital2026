package backend

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateTask(t *testing.T) {
	handler := NewHandler(NewStore())

	req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewBufferString(`{"title":"Новая задача"}`))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("ожидался статус 201, получен %d", rec.Code)
	}
}
