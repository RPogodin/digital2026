package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

const (
	serverAddr = ":18085"
	dbPath     = "tasks.db"
)

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

type app struct {
	db *sql.DB
}

func main() {
	// Здесь меняется только хранилище: HTTP-часть почти такая же, как на этапе 04.
	db, err := openDB(dbPath)
	if err != nil {
		log.Fatalf("[этап 05] не удалось открыть базу данных: %v", err)
	}
	defer db.Close()

	server := &app{db: db}

	log.Printf("[этап 05] база данных готова: %s", dbPath)
	log.Printf("[этап 05] запуск HTTP-сервера на %s", serverAddr)

	if err := http.ListenAndServe(serverAddr, server.routes()); err != nil {
		log.Fatalf("[этап 05] сервер завершился с ошибкой: %v", err)
	}
}

func openDB(path string) (*sql.DB, error) {
	// Файл SQLite и таблица создаются автоматически, чтобы этап запускался без подготовки.
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		log.Printf("[этап 05] файл базы данных еще не существует, он будет создан автоматически")
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	schema := `
	CREATE TABLE IF NOT EXISTS items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		done INTEGER NOT NULL DEFAULT 0,
		created_at TEXT NOT NULL
	);
	`

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, err
	}

	if err := seedDemoItems(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func seedDemoItems(db *sql.DB) error {
	// Стартовые записи нужны, чтобы после первого запуска сразу было что показывать в curl.
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM items`).Scan(&count); err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	log.Printf("[этап 05] таблица пуста, добавляю стартовые данные для демонстрации")

	demoItems := []itemInput{
		{Title: "Посмотреть список задач из SQLite", Done: true},
		{Title: "Создать новую задачу через POST", Done: false},
	}

	for _, current := range demoItems {
		createdAt := time.Now().UTC().Format(time.RFC3339)
		if _, err := db.Exec(`
			INSERT INTO items (title, done, created_at)
			VALUES (?, ?, ?)
		`, current.Title, boolToInt(current.Done), createdAt); err != nil {
			return err
		}
	}

	return nil
}

func (a *app) routes() http.Handler {
	// Набор маршрутов специально оставлен почти таким же, как на этапе 04.
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/items", a.handleList)
	mux.HandleFunc("POST /api/items", a.handleCreate)
	mux.HandleFunc("GET /api/items/{id}", a.handleGet)
	mux.HandleFunc("PUT /api/items/{id}", a.handleUpdate)
	mux.HandleFunc("DELETE /api/items/{id}", a.handleDelete)
	return mux
}

func (a *app) handleList(w http.ResponseWriter, r *http.Request) {
	log.Printf("[этап 05] GET /api/items")

	items, err := a.listItems()
	if err != nil {
		writeInternalError(w, err)
		return
	}

	log.Printf("[этап 05] список элементов подготовлен: count=%d", len(items))
	writeJSON(w, http.StatusOK, items)
}

func (a *app) handleGet(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	log.Printf("[этап 05] GET /api/items/%d", id)

	current, err := a.getItem(id)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "элемент не найден")
		return
	}
	if err != nil {
		writeInternalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, current)
}

func (a *app) handleCreate(w http.ResponseWriter, r *http.Request) {
	log.Printf("[этап 05] POST /api/items")

	input, err := decodeAndValidateInput(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	current, err := a.createItem(input)
	if err != nil {
		writeInternalError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, current)
}

func (a *app) handleUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	log.Printf("[этап 05] PUT /api/items/%d", id)

	input, err := decodeAndValidateInput(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	current, err := a.updateItem(id, input)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "элемент не найден")
		return
	}
	if err != nil {
		writeInternalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, current)
}

func (a *app) handleDelete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	log.Printf("[этап 05] DELETE /api/items/%d", id)

	deleted, err := a.deleteItem(id)
	if err != nil {
		writeInternalError(w, err)
		return
	}
	if !deleted {
		writeError(w, http.StatusNotFound, "элемент не найден")
		return
	}

	log.Printf("[этап 05] ответ без тела отправлен: status=%d", http.StatusNoContent)
	w.WriteHeader(http.StatusNoContent)
}

func (a *app) listItems() ([]item, error) {
	// Здесь удобно ставить breakpoint и показывать живой SQL SELECT.
	log.Printf("[этап 05] SQL SELECT: читаю список элементов")

	rows, err := a.db.Query(`
		SELECT id, title, done, created_at
		FROM items
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []item
	for rows.Next() {
		var current item
		var createdAt string
		var done int

		if err := rows.Scan(&current.ID, &current.Title, &done, &createdAt); err != nil {
			return nil, err
		}

		current.Done = done == 1
		current.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
		if err != nil {
			return nil, err
		}

		items = append(items, current)
	}

	return items, rows.Err()
}

func (a *app) getItem(id int) (item, error) {
	// Чтение одной записи по id оформлено отдельно для наглядности.
	log.Printf("[этап 05] SQL SELECT: читаю элемент id=%d", id)

	var current item
	var done int
	var createdAt string

	err := a.db.QueryRow(`
		SELECT id, title, done, created_at
		FROM items
		WHERE id = ?
	`, id).Scan(&current.ID, &current.Title, &done, &createdAt)
	if err != nil {
		return item{}, err
	}

	current.Done = done == 1
	current.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return item{}, err
	}

	return current, nil
}

func (a *app) createItem(input itemInput) (item, error) {
	// После INSERT сразу читаем запись обратно, чтобы вернуть клиенту полный объект.
	createdAt := time.Now().UTC().Format(time.RFC3339)

	result, err := a.db.Exec(`
		INSERT INTO items (title, done, created_at)
		VALUES (?, ?, ?)
	`, input.Title, boolToInt(input.Done), createdAt)
	if err != nil {
		return item{}, err
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		return item{}, err
	}

	log.Printf("[этап 05] SQL INSERT выполнен: id=%d title=%q", lastID, input.Title)

	return a.getItem(int(lastID))
}

func (a *app) updateItem(id int, input itemInput) (item, error) {
	// UPDATE показывает, как SQL меняет существующую запись.
	result, err := a.db.Exec(`
		UPDATE items
		SET title = ?, done = ?
		WHERE id = ?
	`, input.Title, boolToInt(input.Done), id)
	if err != nil {
		return item{}, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return item{}, err
	}
	if affected == 0 {
		return item{}, sql.ErrNoRows
	}

	log.Printf("[этап 05] SQL UPDATE выполнен: id=%d title=%q done=%t", id, input.Title, input.Done)

	return a.getItem(id)
}

func (a *app) deleteItem(id int) (bool, error) {
	// DELETE возвращает флаг, была ли запись реально найдена.
	result, err := a.db.Exec(`
		DELETE FROM items
		WHERE id = ?
	`, id)
	if err != nil {
		return false, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	if affected == 0 {
		return false, nil
	}

	log.Printf("[этап 05] SQL DELETE выполнен: id=%d", id)
	return true, nil
}

func decodeAndValidateInput(r *http.Request) (itemInput, error) {
	defer r.Body.Close()

	// Разбор JSON и валидация собраны в одном месте, чтобы handler оставался коротким.
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
	// Это учебная функция: она делает ошибки JSON явно читаемыми в терминале и curl.
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

func boolToInt(value bool) int {
	if value {
		return 1
	}

	return 0
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("[этап 05] ошибка записи JSON-ответа: %v", err)
		return
	}

	log.Printf("[этап 05] JSON-ответ отправлен: status=%d", status)
}

func writeError(w http.ResponseWriter, status int, message string) {
	log.Printf("[этап 05] ошибка запроса: status=%d message=%s", status, message)
	writeJSON(w, status, map[string]string{"error": message})
}

func writeInternalError(w http.ResponseWriter, err error) {
	log.Printf("[этап 05] внутренняя ошибка: %v", err)
	writeError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
}
