package backend

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// NewHandler собирает все HTTP-маршруты финального API.
func NewHandler(store *Store) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/tasks", func(w http.ResponseWriter, r *http.Request) {
		handleList(store, w, r)
	})
	mux.HandleFunc("GET /api/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		handleGet(store, w, r)
	})
	mux.HandleFunc("POST /api/tasks", func(w http.ResponseWriter, r *http.Request) {
		handleCreate(store, w, r)
	})
	mux.HandleFunc("PUT /api/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		handleUpdate(store, w, r)
	})
	mux.HandleFunc("DELETE /api/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		handleDelete(store, w, r)
	})
	return mux
}

// handleList показывает разбор query-параметров и чтение списка из SQLite.
func handleList(store *Store, w http.ResponseWriter, r *http.Request) {
	filter := TaskFilter{
		Query:    strings.TrimSpace(r.URL.Query().Get("q")),
		Status:   strings.TrimSpace(r.URL.Query().Get("status")),
		Category: strings.TrimSpace(r.URL.Query().Get("category")),
	}

	log.Printf("[этап 07] API: GET /api/tasks q=%q status=%q category=%q", filter.Query, filter.Status, filter.Category)

	if err := validateTaskFilter(filter); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	tasks, err := store.List(filter)
	if err != nil {
		writeInternalError(w, err)
		return
	}

	log.Printf("[этап 07] API: список задач подготовлен count=%d", len(tasks))
	writeJSON(w, http.StatusOK, tasks)
}

// handleGet читает id из URL и загружает одну задачу.
func handleGet(store *Store, w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	log.Printf("[этап 07] API: GET /api/tasks/%d", id)

	task, err := store.Get(id)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "задача не найдена")
		return
	}
	if err != nil {
		writeInternalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, task)
}

// handleCreate принимает JSON и создает новую задачу.
func handleCreate(store *Store, w http.ResponseWriter, r *http.Request) {
	log.Printf("[этап 07] API: POST /api/tasks")

	input, err := decodeTaskInput(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	task, err := store.Create(input)
	if err != nil {
		writeInternalError(w, err)
		return
	}

	log.Printf("[этап 07] API: задача создана id=%d", task.ID)
	writeJSON(w, http.StatusCreated, task)
}

// handleUpdate показывает путь обновления существующей записи.
func handleUpdate(store *Store, w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	log.Printf("[этап 07] API: PUT /api/tasks/%d", id)

	input, err := decodeTaskInput(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	task, err := store.Update(id, input)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "задача не найдена")
		return
	}
	if err != nil {
		writeInternalError(w, err)
		return
	}

	log.Printf("[этап 07] API: задача обновлена id=%d", task.ID)
	writeJSON(w, http.StatusOK, task)
}

// handleDelete удаляет задачу по id.
func handleDelete(store *Store, w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	log.Printf("[этап 07] API: DELETE /api/tasks/%d", id)

	deleted, err := store.Delete(id)
	if err != nil {
		writeInternalError(w, err)
		return
	}
	if !deleted {
		writeError(w, http.StatusNotFound, "задача не найдена")
		return
	}

	log.Printf("[этап 07] API: ответ без тела отправлен status=%d", http.StatusNoContent)
	w.WriteHeader(http.StatusNoContent)
}

// decodeTaskInput разбирает JSON и одновременно показывает типичную валидацию API.
func decodeTaskInput(r *http.Request) (TaskInput, error) {
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var input TaskInput
	if err := decoder.Decode(&input); err != nil {
		return TaskInput{}, explainJSONError(err)
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return TaskInput{}, errors.New("тело запроса должно содержать только один JSON-объект")
	}

	if err := ValidateTaskInput(input); err != nil {
		return TaskInput{}, err
	}

	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	return input, nil
}

// explainJSONError превращает техническую ошибку декодера в сообщение для лекции и клиента.
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

// validateTaskFilter проверяет query-параметры списка.
func validateTaskFilter(filter TaskFilter) error {
	if filter.Status != "" && !validStatuses[filter.Status] {
		return errors.New("параметр status должен быть одним из: todo, doing, done")
	}

	if filter.Category != "" && !validCategories[filter.Category] {
		return errors.New("параметр category должен быть одним из: lecture, demo, infra")
	}

	return nil
}

// parseID читает параметр маршрута /api/tasks/{id}.
func parseID(r *http.Request) (int, error) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		return 0, errors.New("id должен быть положительным числом")
	}

	return id, nil
}

// writeJSON централизует запись JSON-ответа и логирует статус.
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("[этап 07] API: ошибка записи JSON-ответа: %v", err)
		return
	}

	log.Printf("[этап 07] API: JSON-ответ отправлен status=%d", status)
}

// writeError отдает клиенту ошибку в едином формате {"error":"..."}.
func writeError(w http.ResponseWriter, status int, message string) {
	log.Printf("[этап 07] API: ошибка запроса status=%d message=%s", status, message)
	writeJSON(w, status, map[string]string{"error": message})
}

func writeInternalError(w http.ResponseWriter, err error) {
	log.Printf("[этап 07] API: внутренняя ошибка: %v", err)
	writeError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
}
