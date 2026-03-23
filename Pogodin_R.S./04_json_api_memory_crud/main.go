package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const serverAddr = ":18084"

type item struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
}

type itemInput struct {
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

type memoryStore struct {
	mu     sync.Mutex
	nextID int
	items  map[int]item
}

func newMemoryStore() *memoryStore {
	store := &memoryStore{
		nextID: 3,
		items: map[int]item{
			1: {
				ID:        1,
				Title:     "Показать список задач",
				Done:      true,
				CreatedAt: time.Now().Add(-20 * time.Minute),
			},
			2: {
				ID:        2,
				Title:     "Создать новую задачу через POST",
				Done:      false,
				CreatedAt: time.Now().Add(-10 * time.Minute),
			},
		},
	}

	return store
}

func main() {
	// Данные пока живут только в памяти процесса.
	store := newMemoryStore()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/items", store.handleList)
	mux.HandleFunc("POST /api/items", store.handleCreate)
	mux.HandleFunc("GET /api/items/{id}", store.handleGet)
	mux.HandleFunc("PUT /api/items/{id}", store.handleUpdate)
	mux.HandleFunc("DELETE /api/items/{id}", store.handleDelete)

	log.Printf("[этап 04] запуск JSON API на %s", serverAddr)

	if err := http.ListenAndServe(serverAddr, mux); err != nil {
		log.Fatalf("[этап 04] сервер завершился с ошибкой: %v", err)
	}
}

func (s *memoryStore) handleList(w http.ResponseWriter, r *http.Request) {
	// Самый простой API-сценарий: просто читаем список и отдаем JSON.
	log.Printf("[этап 04] GET /api/items")
	items := s.list()
	log.Printf("[этап 04] список элементов подготовлен: count=%d", len(items))
	writeJSON(w, http.StatusOK, items)
}

func (s *memoryStore) handleGet(w http.ResponseWriter, r *http.Request) {
	// Здесь хорошо видно, как id достается из маршрута /api/items/{id}.
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	log.Printf("[этап 04] GET /api/items/%d", id)

	item, ok := s.get(id)
	if !ok {
		writeError(w, http.StatusNotFound, "элемент не найден")
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func (s *memoryStore) handleCreate(w http.ResponseWriter, r *http.Request) {
	// POST показывает полный путь: JSON -> Go-структура -> бизнес-логика -> JSON-ответ.
	log.Printf("[этап 04] POST /api/items")

	input, err := decodeAndValidateInput(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	item := s.create(input)
	writeJSON(w, http.StatusCreated, item)
}

func (s *memoryStore) handleUpdate(w http.ResponseWriter, r *http.Request) {
	// PUT почти повторяет POST, но добавляет поиск существующей записи по id.
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	log.Printf("[этап 04] PUT /api/items/%d", id)

	input, err := decodeAndValidateInput(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	item, ok := s.update(id, input)
	if !ok {
		writeError(w, http.StatusNotFound, "элемент не найден")
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func (s *memoryStore) handleDelete(w http.ResponseWriter, r *http.Request) {
	// DELETE удобно показывать как пример ответа без тела.
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	log.Printf("[этап 04] DELETE /api/items/%d", id)

	if !s.delete(id) {
		writeError(w, http.StatusNotFound, "элемент не найден")
		return
	}

	log.Printf("[этап 04] ответ без тела отправлен: status=%d", http.StatusNoContent)
	w.WriteHeader(http.StatusNoContent)
}

func (s *memoryStore) list() []item {
	s.mu.Lock()
	defer s.mu.Unlock()

	items := make([]item, 0, len(s.items))
	for _, current := range s.items {
		items = append(items, current)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})

	return items
}

func (s *memoryStore) get(id int) (item, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	current, ok := s.items[id]
	return current, ok
}

func (s *memoryStore) create(input itemInput) item {
	s.mu.Lock()
	defer s.mu.Unlock()

	current := item{
		ID:        s.nextID,
		Title:     input.Title,
		Done:      input.Done,
		CreatedAt: time.Now(),
	}

	s.items[current.ID] = current
	s.nextID++

	log.Printf("[этап 04] создан новый элемент: id=%d title=%q", current.ID, current.Title)

	return current
}

func (s *memoryStore) update(id int, input itemInput) (item, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	current, ok := s.items[id]
	if !ok {
		return item{}, false
	}

	current.Title = input.Title
	current.Done = input.Done
	s.items[id] = current

	log.Printf("[этап 04] элемент обновлен: id=%d title=%q done=%t", current.ID, current.Title, current.Done)

	return current, true
}

func (s *memoryStore) delete(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.items[id]; !ok {
		return false
	}

	delete(s.items, id)
	log.Printf("[этап 04] элемент удален: id=%d", id)
	return true
}

func decodeAndValidateInput(r *http.Request) (itemInput, error) {
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var input itemInput
	if err := decoder.Decode(&input); err != nil {
		return itemInput{}, explainJSONError(err)
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return itemInput{}, errors.New("тело запроса должно содержать только один JSON-объект")
	}

	input.Title = strings.TrimSpace(input.Title)
	if input.Title == "" {
		return itemInput{}, errors.New("поле title обязательно")
	}

	if len(input.Title) > 120 {
		return itemInput{}, errors.New("поле title слишком длинное")
	}

	return input, nil
}

func explainJSONError(err error) error {
	var syntaxError *json.SyntaxError
	var typeError *json.UnmarshalTypeError

	switch {
	case errors.Is(err, io.EOF):
		return errors.New("тело запроса не должно быть пустым")
	case errors.As(err, &syntaxError):
		return fmt.Errorf("в JSON есть синтаксическая ошибка около позиции %d", syntaxError.Offset)
	case errors.As(err, &typeError):
		if typeError.Field == "" {
			return errors.New("в JSON передан неверный тип данных")
		}

		return fmt.Errorf("поле %s имеет неверный тип", typeError.Field)
	case strings.HasPrefix(err.Error(), "json: unknown field "):
		return fmt.Errorf("в JSON передано неизвестное поле %s", strings.TrimPrefix(err.Error(), "json: unknown field "))
	default:
		return errors.New("тело запроса должно быть корректным JSON")
	}
}

func parseID(r *http.Request) (int, error) {
	rawID := r.PathValue("id")
	id, err := strconv.Atoi(rawID)
	if err != nil || id <= 0 {
		return 0, errors.New("id должен быть положительным числом")
	}

	return id, nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("[этап 04] ошибка записи JSON-ответа: %v", err)
		return
	}

	log.Printf("[этап 04] JSON-ответ отправлен: status=%d", status)
}

func writeError(w http.ResponseWriter, status int, message string) {
	log.Printf("[этап 04] ошибка запроса: status=%d message=%s", status, message)
	writeJSON(w, status, map[string]string{"error": message})
}
