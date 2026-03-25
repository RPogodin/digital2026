package main

import (
	"fmt"
	"log"
	"net/http"
)

const serverAddr = ":18081"

func main() {
	log.Printf("[этап 01] запуск HTTP-сервера на %s", serverAddr)

	// На первом этапе достаточно одного маршрута и одного handler.
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", helloHandler)

	if err := http.ListenAndServe(serverAddr, mux); err != nil {
		log.Fatalf("[этап 01] сервер завершился с ошибкой: %v", err)
	}
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	// Здесь удобно ставить первый breakpoint: запрос уже пришел в Go-код.
	log.Printf("[этап 01] пришел запрос: method=%s path=%s remote=%s", r.Method, r.URL.Path, r.RemoteAddr)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	// Ответ формируется буквально одной записью в ResponseWriter.
	if _, err := fmt.Fprintln(w, "Hello, HTTP!"); err != nil {
		log.Printf("[этап 01] не удалось отправить ответ: %v", err)
		return
	}

	log.Printf("[этап 01] ответ успешно отправлен")
}
