package main

import (
	"embed"
	"html/template"
	"log"
	"net/http"
	"strings"
)

const serverAddr = ":18083"

//go:embed templates/*.html
var templateFS embed.FS

type pageData struct {
	QueryName   string
	FormName    string
	Message     string
	Method      string
	RequestPath string
}

func main() {
	// Шаблон страницы с формой подготавливаем заранее.
	tmpl, err := template.ParseFS(templateFS, "templates/index.html")
	if err != nil {
		log.Fatalf("[этап 03] не удалось разобрать HTML-шаблон: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		// GET показывает чтение query-параметров из URL.
		log.Printf("[этап 03] GET /: path=%s raw_query=%s", r.URL.Path, r.URL.RawQuery)

		data := pageData{
			QueryName:   strings.TrimSpace(r.URL.Query().Get("name")),
			Method:      r.Method,
			RequestPath: r.URL.Path,
		}

		if data.QueryName != "" {
			data.Message = "Привет из query string, " + data.QueryName + "!"
		}

		renderPage(w, tmpl, data)
	})

	mux.HandleFunc("POST /hello", func(w http.ResponseWriter, r *http.Request) {
		// POST показывает чтение данных формы из тела запроса.
		log.Printf("[этап 03] POST /hello: path=%s", r.URL.Path)

		if err := r.ParseForm(); err != nil {
			log.Printf("[этап 03] ошибка разбора формы: %v", err)
			http.Error(w, "не удалось разобрать форму", http.StatusBadRequest)
			return
		}

		formName := strings.TrimSpace(r.FormValue("name"))
		if formName == "" {
			formName = "гость"
		}

		data := pageData{
			FormName:    formName,
			Message:     "Привет из POST-формы, " + formName + "!",
			Method:      r.Method,
			RequestPath: r.URL.Path,
		}

		renderPage(w, tmpl, data)
	})

	log.Printf("[этап 03] запуск HTTP-сервера на %s", serverAddr)

	if err := http.ListenAndServe(serverAddr, mux); err != nil {
		log.Fatalf("[этап 03] сервер завершился с ошибкой: %v", err)
	}
}

func renderPage(w http.ResponseWriter, tmpl *template.Template, data pageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Общая точка рендеринга удобна для breakpoint и логирования готового ответа.
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("[этап 03] ошибка выполнения шаблона: %v", err)
		http.Error(w, "не удалось построить страницу", http.StatusInternalServerError)
		return
	}

	log.Printf("[этап 03] страница успешно отправлена: method=%s path=%s", data.Method, data.RequestPath)
}
