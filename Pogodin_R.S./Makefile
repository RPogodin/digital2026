STAGES := 01 02 03 04 05 06 07

.PHONY: help $(addprefix run-,$(STAGES)) test fmt

help:
	@printf "Доступные команды:\n"
	@printf "  make run-01  # Этап 01: Hello HTTP\n"
	@printf "  make run-02  # Этап 02: HTML-ответ\n"
	@printf "  make run-03  # Этап 03: Формы и параметры\n"
	@printf "  make run-04  # Этап 04: JSON API в памяти\n"
	@printf "  make run-05  # Этап 05: CRUD с SQLite\n"
	@printf "  make run-06  # Этап 06: Frontend + Backend\n"
	@printf "  make run-07  # Этап 07: Fullstack-приложение\n"
	@printf "  make test    # Запуск тестов\n"
	@printf "  make fmt     # Форматирование Go-кода\n"

run-01:
	cd 01_hello_http && go run .

run-02:
	cd 02_html_response && go run .

run-03:
	cd 03_form_and_params && go run .

run-04:
	cd 04_json_api_memory_crud && go run .

run-05:
	cd 05_crud_with_sqlite && go run .

run-06:
	cd 06_frontend_backend_split && go run .

run-07:
	cd 07_fullstack_app && go run .

test:
	go test ./...

fmt:
	gofmt -w $(shell find . -name '*.go' -type f)

