package backend

import (
	"database/sql"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type Task struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Category    string    `json:"category"`
	CreatedAt   time.Time `json:"created_at"`
}

type TaskInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Category    string `json:"category"`
}

type TaskFilter struct {
	Query    string
	Status   string
	Category string
}

var (
	validStatuses = map[string]bool{
		"todo":  true,
		"doing": true,
		"done":  true,
	}
	validCategories = map[string]bool{
		"lecture": true,
		"demo":    true,
		"infra":   true,
	}
)

type Store struct {
	db *sql.DB
}

// NewStore открывает SQLite и подготавливает таблицу.
func NewStore(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	schema := `
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		description TEXT NOT NULL,
		status TEXT NOT NULL,
		category TEXT NOT NULL,
		created_at TEXT NOT NULL
	);
	`

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, err
	}

	if err := seedDemoTasks(db); err != nil {
		db.Close()
		return nil, err
	}

	return &Store{db: db}, nil
}

// seedDemoTasks заполняет пустую таблицу стартовыми данными для лекции.
func seedDemoTasks(db *sql.DB) error {
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM tasks`).Scan(&count); err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	log.Printf("[этап 07] store: таблица пуста, добавляю стартовые задачи для демонстрации")

	demoTasks := []TaskInput{
		{
			Title:       "Подготовить сценарий лекции",
			Description: "Показать путь от клика в браузере до SQL-запроса",
			Status:      "todo",
			Category:    "lecture",
		},
		{
			Title:       "Проверить DevTools",
			Description: "Открыть вкладки Network, Console и Elements",
			Status:      "doing",
			Category:    "demo",
		},
		{
			Title:       "Почистить временную базу",
			Description: "Показать, где лежит файл SQLite и как его пересоздать",
			Status:      "done",
			Category:    "infra",
		},
	}

	for _, current := range demoTasks {
		createdAt := time.Now().UTC().Format(time.RFC3339)
		if _, err := db.Exec(`
			INSERT INTO tasks (title, description, status, category, created_at)
			VALUES (?, ?, ?, ?, ?)
		`, current.Title, current.Description, current.Status, current.Category, createdAt); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

// ValidateTaskInput проверяет поля, которые приходят из HTTP-запроса.
func ValidateTaskInput(input TaskInput) error {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return errors.New("поле title обязательно")
	}

	if len(input.Title) > 120 {
		return errors.New("поле title слишком длинное")
	}

	if len(input.Description) > 500 {
		return errors.New("поле description слишком длинное")
	}

	if !validStatuses[input.Status] {
		return errors.New("поле status должно быть одним из: todo, doing, done")
	}

	if !validCategories[input.Category] {
		return errors.New("поле category должно быть одним из: lecture, demo, infra")
	}

	return nil
}

// List строит SQL с простыми фильтрами и возвращает список задач.
func (s *Store) List(filter TaskFilter) ([]Task, error) {
	log.Printf(
		"[этап 07] store: SQL SELECT список задач q=%q status=%q category=%q",
		filter.Query,
		filter.Status,
		filter.Category,
	)

	query := `
		SELECT id, title, description, status, category, created_at
		FROM tasks
		WHERE 1 = 1
	`

	var args []any

	if filter.Query != "" {
		query += ` AND (title LIKE ? OR description LIKE ?)`
		mask := "%" + filter.Query + "%"
		args = append(args, mask, mask)
	}

	if filter.Status != "" {
		query += ` AND status = ?`
		args = append(args, filter.Status)
	}

	if filter.Category != "" {
		query += ` AND category = ?`
		args = append(args, filter.Category)
	}

	query += ` ORDER BY created_at DESC, id DESC`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		var createdAt string

		if err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.Category,
			&createdAt,
		); err != nil {
			return nil, err
		}

		task.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

// Get читает одну задачу по id.
func (s *Store) Get(id int) (Task, error) {
	log.Printf("[этап 07] store: SQL SELECT задача id=%d", id)

	row := s.db.QueryRow(`
		SELECT id, title, description, status, category, created_at
		FROM tasks
		WHERE id = ?
	`, id)

	var task Task
	var createdAt string

	err := row.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.Category,
		&createdAt,
	)
	if err != nil {
		return Task{}, err
	}

	task.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return Task{}, err
	}

	return task, nil
}

// Create вставляет новую задачу и потом читает ее обратно.
func (s *Store) Create(input TaskInput) (Task, error) {
	createdAt := time.Now().UTC().Format(time.RFC3339)

	log.Printf(
		"[этап 07] store: SQL INSERT title=%q status=%s category=%s",
		input.Title,
		input.Status,
		input.Category,
	)

	result, err := s.db.Exec(`
		INSERT INTO tasks (title, description, status, category, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, strings.TrimSpace(input.Title), strings.TrimSpace(input.Description), input.Status, input.Category, createdAt)
	if err != nil {
		return Task{}, err
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		return Task{}, err
	}

	return s.Get(int(lastID))
}

// Update изменяет существующую задачу.
func (s *Store) Update(id int, input TaskInput) (Task, error) {
	log.Printf(
		"[этап 07] store: SQL UPDATE id=%d title=%q status=%s category=%s",
		id,
		input.Title,
		input.Status,
		input.Category,
	)

	result, err := s.db.Exec(`
		UPDATE tasks
		SET title = ?, description = ?, status = ?, category = ?
		WHERE id = ?
	`, strings.TrimSpace(input.Title), strings.TrimSpace(input.Description), input.Status, input.Category, id)
	if err != nil {
		return Task{}, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return Task{}, err
	}
	if affected == 0 {
		return Task{}, sql.ErrNoRows
	}

	return s.Get(id)
}

// Delete удаляет задачу и возвращает, была ли она найдена.
func (s *Store) Delete(id int) (bool, error) {
	log.Printf("[этап 07] store: SQL DELETE id=%d", id)

	result, err := s.db.Exec(`DELETE FROM tasks WHERE id = ?`, id)
	if err != nil {
		return false, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return affected > 0, nil
}
