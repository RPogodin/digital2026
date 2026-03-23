package backend

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

type Task struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
}

type taskInput struct {
	Title string `json:"title"`
}

type Store struct {
	mu     sync.Mutex
	nextID int
	tasks  map[int]Task
}

// NewStore создает простое in-memory хранилище со стартовыми данными.
func NewStore() *Store {
	return &Store{
		nextID: 3,
		tasks: map[int]Task{
			1: {ID: 1, Title: "Открыть страницу в браузере", CreatedAt: time.Now().Add(-20 * time.Minute)},
			2: {ID: 2, Title: "Создать задачу через fetch", CreatedAt: time.Now().Add(-10 * time.Minute)},
		},
	}
}

// NewHandler собирает маленький JSON API для frontend.
func NewHandler(store *Store) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/tasks", store.handleList)
	mux.HandleFunc("POST /api/tasks", store.handleCreate)
	mux.HandleFunc("DELETE /api/tasks/{id}", store.handleDelete)
	return mux
}

func (s *Store) handleList(w http.ResponseWriter, r *http.Request) {
	// GET /api/tasks нужен frontend для начальной загрузки и обновления списка.
	log.Printf("[этап 06] backend: GET /api/tasks")
	tasks := s.list()
	log.Printf("[этап 06] backend: список задач подготовлен count=%d", len(tasks))
	writeJSON(w, http.StatusOK, tasks)
}

func (s *Store) handleCreate(w http.ResponseWriter, r *http.Request) {
	// POST /api/tasks принимает JSON из fetch.
	log.Printf("[этап 06] backend: POST /api/tasks")

	input, err := decodeTaskInput(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	task := s.create(input.Title)
	writeJSON(w, http.StatusCreated, task)
}

func (s *Store) handleDelete(w http.ResponseWriter, r *http.Request) {
	// DELETE удобно показывать как короткий сценарий без ответа в теле.
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	log.Printf("[этап 06] backend: DELETE /api/tasks/%d", id)

	if !s.delete(id) {
		writeError(w, http.StatusNotFound, "задача не найдена")
		return
	}

	log.Printf("[этап 06] backend: ответ без тела отправлен status=%d", http.StatusNoContent)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Store) list() []Task {
	// Копируем данные в срез, чтобы отдать их в стабильном порядке.
	s.mu.Lock()
	defer s.mu.Unlock()

	tasks := make([]Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].ID < tasks[j].ID
	})

	return tasks
}

func (s *Store) create(title string) Task {
	// Создание новой задачи происходит прямо в памяти процесса.
	s.mu.Lock()
	defer s.mu.Unlock()

	task := Task{
		ID:        s.nextID,
		Title:     title,
		CreatedAt: time.Now(),
	}

	s.tasks[task.ID] = task
	s.nextID++

	log.Printf("[этап 06] backend: задача создана id=%d title=%q", task.ID, task.Title)

	return task
}

func (s *Store) delete(id int) bool {
	// Удаление возвращает false, если задачи с таким id не было.
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.tasks[id]; !ok {
		return false
	}

	delete(s.tasks, id)
	log.Printf("[этап 06] backend: задача удалена id=%d", id)
	return true
}

func decodeTaskInput(r *http.Request) (taskInput, error) {
	defer r.Body.Close()

	// Проверяем JSON строго, чтобы frontend-ошибки были заметны сразу.
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var input taskInput
	if err := decoder.Decode(&input); err != nil {
		return taskInput{}, explainJSONError(err)
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return taskInput{}, errors.New("тело запроса должно содержать только один JSON-объект")
	}

	input.Title = strings.TrimSpace(input.Title)
	if input.Title == "" {
		return taskInput{}, errors.New("поле title обязательно")
	}

	if len(input.Title) > 120 {
		return taskInput{}, errors.New("поле title слишком длинное")
	}

	return input, nil
}

func explainJSONError(err error) error {
	// Ошибки JSON специально переведены на человеческий язык для лекции.
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
	// PathValue читает id из маршрута /api/tasks/{id}.
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		return 0, errors.New("id должен быть положительным числом")
	}

	return id, nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("[этап 06] backend: ошибка записи JSON-ответа: %v", err)
		return
	}

	log.Printf("[этап 06] backend: JSON-ответ отправлен status=%d", status)
}

func writeError(w http.ResponseWriter, status int, message string) {
	log.Printf("[этап 06] backend: ошибка запроса status=%d message=%s", status, message)
	writeJSON(w, status, map[string]string{"error": message})
}
