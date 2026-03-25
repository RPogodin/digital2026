package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"

	"example.com/http-evolution-demo/07_fullstack_app/backend"
)

//go:embed frontend/*
var frontendFS embed.FS

func main() {
	// Конфигурация предельно простая: берем порт и путь к БД из env или используем значения по умолчанию.
	port := envOrDefault("APP_PORT", ":18087")
	dbPath := envOrDefault("APP_DB_PATH", "data/tasks.db")

	// Поднимаем SQLite-хранилище. Если файла БД еще нет, оно само создастся.
	store, err := backend.NewStore(dbPath)
	if err != nil {
		log.Fatalf("[этап 07] не удалось подготовить хранилище: %v", err)
	}
	defer store.Close()

	// API и раздача статики работают на одном HTTP-сервере.
	api := backend.NewHandler(store)

	staticFS, err := fs.Sub(frontendFS, "frontend")
	if err != nil {
		log.Fatalf("[этап 07] не удалось подготовить frontend: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/api/", api)
	mux.Handle("/", http.FileServer(http.FS(staticFS)))

	log.Printf("[этап 07] запуск fullstack-приложения на %s", port)
	log.Printf("[этап 07] файл базы данных: %s", dbPath)
	log.Printf("[этап 07] откройте в браузере: http://localhost%s/", port)

	if err := http.ListenAndServe(port, logRequests(mux)); err != nil {
		log.Fatalf("[этап 07] сервер завершился с ошибкой: %v", err)
	}
}

// logRequests добавляет один прозрачный лог вокруг любого запроса.
func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[этап 07] HTTP: method=%s path=%s", r.Method, r.URL.RequestURI())
		next.ServeHTTP(w, r)
	})
}

func envOrDefault(name string, fallback string) string {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}

	return value
}
