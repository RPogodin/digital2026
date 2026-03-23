package main

import (
	"embed"
	"html/template"
	"log"
	"net/http"
	"time"
)

const serverAddr = ":18082"

//go:embed templates/*.html
var templateFS embed.FS

type pageData struct {
	Title       string
	ServerTime  string
	RequestPath string
}

func main() {
	// Шаблон читается при старте, чтобы на каждом запросе не разбирать его заново.
	tmpl, err := template.ParseFS(templateFS, "templates/index.html")
	if err != nil {
		log.Fatalf("[этап 02] не удалось разобрать HTML-шаблон: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		// На этом этапе поток уже такой: запрос -> структура данных -> шаблон -> HTML.
		log.Printf("[этап 02] пришел запрос: method=%s path=%s", r.Method, r.URL.Path)

		data := pageData{
			Title:       "HTML-ответ от сервера",
			ServerTime:  time.Now().Format("02.01.2006 15:04:05"),
			RequestPath: r.URL.Path,
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		// Здесь удобно смотреть, как данные попадают в HTML-шаблон.
		if err := tmpl.Execute(w, data); err != nil {
			log.Printf("[этап 02] ошибка выполнения шаблона: %v", err)
			http.Error(w, "не удалось построить HTML-страницу", http.StatusInternalServerError)
			return
		}

		log.Printf("[этап 02] HTML-страница успешно отправлена")
	})

	log.Printf("[этап 02] запуск HTTP-сервера на %s", serverAddr)

	if err := http.ListenAndServe(serverAddr, mux); err != nil {
		log.Fatalf("[этап 02] сервер завершился с ошибкой: %v", err)
	}
}
