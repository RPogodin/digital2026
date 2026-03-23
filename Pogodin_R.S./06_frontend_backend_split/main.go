package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"

	"example.com/http-evolution-demo/06_frontend_backend_split/backend"
)

const serverAddr = ":18086"

//go:embed frontend/*
var frontendFS embed.FS

func main() {
	// Backend создается отдельно, чтобы на лекции было видно разделение ролей.
	mux := http.NewServeMux()
	store := backend.NewStore()
	mux.Handle("/api/", backend.NewHandler(store))

	// Статические файлы frontend отдаются тем же Go-сервером.
	staticFS, err := fs.Sub(frontendFS, "frontend")
	if err != nil {
		log.Fatalf("[этап 06] не удалось подготовить frontend: %v", err)
	}

	mux.Handle("/", http.FileServer(http.FS(staticFS)))

	log.Printf("[этап 06] запуск сервера на %s", serverAddr)
	log.Printf("[этап 06] frontend доступен по адресу http://localhost%s/", serverAddr)
	log.Printf("[этап 06] backend API доступен по адресу http://localhost%s/api/tasks", serverAddr)

	if err := http.ListenAndServe(serverAddr, logRequests(mux)); err != nil {
		log.Fatalf("[этап 06] сервер завершился с ошибкой: %v", err)
	}
}

func logRequests(next http.Handler) http.Handler {
	// Один общий лог помогает видеть и загрузку страницы, и API-вызовы.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[этап 06] HTTP: method=%s path=%s", r.Method, r.URL.RequestURI())
		next.ServeHTTP(w, r)
	})
}
