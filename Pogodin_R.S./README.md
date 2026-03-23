# Учебный репозиторий: эволюция backend + frontend по HTTP

## Как установить Go на macOS и Ubuntu

### macOS

Самый простой вариант через Homebrew:

```bash
brew install go
go version
```

Если Homebrew не используется, скачайте официальный `.pkg`-установщик с
<https://go.dev/dl/> и после установки проверьте:

```bash
go version
```

### Ubuntu

Самый простой вариант для Ubuntu:

```bash
sudo apt update
sudo apt install -y golang-go
go version
```

Если нужна более свежая версия Go, чем в репозитории Ubuntu, используйте
официальный архив с <https://go.dev/dl/>:

```bash
wget https://go.dev/dl/go<VERSION>.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go<VERSION>.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.profile
source ~/.profile
go version
```

Если у вас Ubuntu на ARM64, замените `linux-amd64` на `linux-arm64`.

Этот репозиторий показывает семь последовательных учебных приложений на Go.
Идея простая: начать с минимального HTTP-сервера и шаг за шагом дойти до
полноценного fullstack-приложения с SQLite, frontend на чистом JavaScript и
понятным потоком данных.

## Что внутри

| Этап | Каталог | Порт | Что демонстрирует |
| --- | --- | --- | --- |
| 01 | `01_hello_http` | `18081` | Самый простой HTTP-сервер и один handler |
| 02 | `02_html_response` | `18082` | HTML-ответ и шаблоны `html/template` |
| 03 | `03_form_and_params` | `18083` | Query-параметры, формы, `GET` и `POST` |
| 04 | `04_json_api_memory_crud` | `18084` | JSON API и CRUD в памяти |
| 05 | `05_crud_with_sqlite` | `18085` | CRUD с SQLite через `database/sql` |
| 06 | `06_frontend_backend_split` | `18086` | Разделение frontend/backend и `fetch` |
| 07 | `07_fullstack_app` | `18087` | Полноценное учебное fullstack-приложение |

## Быстрый старт

Запуск из корня репозитория:

```bash
make run-01
make run-02
make run-03
make run-04
make run-05
make run-06
make run-07
```

Показать список команд:

```bash
make help
```

Запустить тесты:

```bash
make test
```

## Общая учебная линия

1. `01_hello_http`  
   Видно, что HTTP-сервер принимает запрос и возвращает простой текст.
2. `02_html_response`  
   Тот же сервер уже умеет возвращать HTML и подставлять данные в шаблон.
3. `03_form_and_params`  
   Появляются входные данные: строка запроса, путь, метод и форма с `POST`.
4. `04_json_api_memory_crud`  
   Переход от HTML-страниц к JSON API и CRUD-операциям.
5. `05_crud_with_sqlite`  
   Те же идеи, но с постоянным хранением в SQLite.
6. `06_frontend_backend_split`  
   Сервер отдает статику, а браузер отдельным запросом ходит в API через `fetch`.
7. `07_fullstack_app`  
   Финальный вариант: UI, фильтры, редактирование, SQLite и чуть более цельная структура.

## Как удобно вести живую демонстрацию

- Сначала показывать браузер и `curl` параллельно.
- В каждом этапе обращать внимание на лог сервера.
- На ранних шагах ставить breakpoint прямо в handler.
- На этапах с API дополнительно смотреть заголовки, JSON и HTTP status code.
- На этапах с frontend открывать DevTools, вкладки `Network` и `Console`.
- На этапах с SQLite заходить в код SQL-запросов и смотреть, как данные проходят путь от HTTP до БД и обратно.

## Единый поток данных

Во всех примерах полезно повторять одну и ту же схему:

`браузер/клиент -> HTTP-запрос -> handler -> бизнес-логика -> ответ`

На поздних этапах цепочка расширяется:

`браузер -> JS/fetch -> HTTP-запрос -> handler -> бизнес-логика -> SQL -> SQLite -> JSON -> JS -> DOM`

## Зависимости

Почти все этапы сделаны только на стандартной библиотеке Go.
Дополнительная зависимость появляется только там, где нужен pure Go SQLite-драйвер,
чтобы не зависеть от CGO и внешних системных библиотек.

## Рекомендуемый порядок демонстрации на лекции

1. Показать `01_hello_http` и зафиксировать базовую схему HTTP: запрос, handler, ответ.
2. Перейти к `02_html_response` и показать, как простой текст превращается в HTML.
3. На `03_form_and_params` разобрать входные данные: URL, query-параметры, метод, форма.
4. На `04_json_api_memory_crud` сделать переход от HTML к JSON API и CRUD.
5. На `05_crud_with_sqlite` показать, что тот же API теперь пишет в постоянное хранилище.
6. На `06_frontend_backend_split` объяснить разделение ролей браузера и сервера.
7. На `07_fullstack_app` собрать все вместе: UI, fetch, валидацию, SQLite, фильтры и редактирование.

## Что именно показывать студентам на каждом этапе

### Этап 01

- `main.go` с `http.ListenAndServe`, `ServeMux` и одним handler.
- `curl -i http://localhost:18081/`.
- Лог сервера при входящем запросе.

### Этап 02

- HTML-шаблон и структуру `pageData`.
- Разницу между обычным текстом и HTML-ответом.
- Как данные из Go попадают в шаблон.

### Этап 03

- Разницу между `GET` и `POST`.
- `r.URL.Query()` и `r.ParseForm()`.
- Форму в браузере и тот же сценарий через `curl`.

### Этап 04

- JSON-запрос и JSON-ответ.
- CRUD-маршруты и HTTP status code.
- Разбор `id` из URL и валидацию входного JSON.

### Этап 05

- SQL `SELECT`, `INSERT`, `UPDATE`, `DELETE` прямо в Go-коде.
- Автоматическое создание файла SQLite и стартовых данных.
- Повторный запуск сервера с сохранением данных.

### Этап 06

- Каталоги `frontend/` и `backend/`.
- `fetch` в `app.js`.
- Как список задач запрашивается и перерисовывается в DOM.

### Этап 07

- Полный поток: форма -> JS -> fetch -> handler -> SQL -> SQLite -> JSON -> DOM.
- Фильтрацию и поиск через query-параметры.
- Редактирование, удаление и общий HTTP-лог.

## Где смотреть сеть / запросы / ответы / логи / базу

- Сеть и ответы в браузере: DevTools -> `Network`.
- Ошибки JavaScript и диагностические сообщения: DevTools -> `Console`.
- Изменение DOM: DevTools -> `Elements`.
- Серверные логи: терминал, где запущен `go run .` или `make run-0X`.
- Файл SQLite на этапе 05: `05_crud_with_sqlite/tasks.db`.
- Файл SQLite на этапе 07: `07_fullstack_app/data/tasks.db`.

Если на машине есть утилита `sqlite3`, базу удобно показать так:

```bash
sqlite3 05_crud_with_sqlite/tasks.db
sqlite3 07_fullstack_app/data/tasks.db
```

Например:

```sql
SELECT * FROM items;
SELECT * FROM tasks;
```
