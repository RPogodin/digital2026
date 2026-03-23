# Этап 04. JSON API + CRUD в памяти

## Учебная цель

Перейти от HTML-страниц к API, которое работает с JSON и поддерживает CRUD.

Поток данных:

`клиент -> HTTP-запрос -> handler -> бизнес-логика -> JSON-ответ`

## Как запустить

Из каталога этапа:

```bash
go run .
```

Из корня репозитория:

```bash
make run-04
```

## Какие URL использовать

- `http://localhost:18084/api/items`
- `http://localhost:18084/api/items/1`

## Что такое JSON API

JSON API — это сервер, который принимает и отдает данные в формате JSON.
Обычно такой API вызывают:

- `curl`;
- frontend-код через `fetch`;
- другие сервисы.

В этом этапе API работает без базы данных: данные лежат в памяти процесса.
После перезапуска сервера они сбрасываются.

## Как выглядит CRUD

- `GET /api/items` — получить список;
- `GET /api/items/{id}` — получить один элемент;
- `POST /api/items` — создать новый элемент;
- `PUT /api/items/{id}` — обновить существующий элемент;
- `DELETE /api/items/{id}` — удалить элемент.

## Как работает сериализация и десериализация в Go

- `json.NewDecoder(r.Body).Decode(&input)` превращает JSON из запроса в Go-структуру;
- `json.NewEncoder(w).Encode(value)` превращает Go-структуру в JSON-ответ.

Это и есть переход между HTTP-данными и типами Go.

## Валидация входных данных

В примере проверяется:

- что тело запроса содержит корректный JSON;
- что в JSON нет лишних полей;
- что поле `title` не пустое;
- что `title` не слишком длинное;
- что `id` в URL является положительным числом.

## Примеры curl

Получить список:

```bash
curl -i http://localhost:18084/api/items
```

Получить один элемент:

```bash
curl -i http://localhost:18084/api/items/1
```

Создать элемент:

```bash
curl -i -X POST http://localhost:18084/api/items \
  -H "Content-Type: application/json" \
  -d '{"title":"Разобрать JSON API","done":false}'
```

Обновить элемент:

```bash
curl -i -X PUT http://localhost:18084/api/items/1 \
  -H "Content-Type: application/json" \
  -d '{"title":"Обновленная задача","done":true}'
```

Удалить элемент:

```bash
curl -i -X DELETE http://localhost:18084/api/items/1
```

## Какие HTTP status code показывает этап

- `200 OK` — успешное чтение и обновление;
- `201 Created` — успешное создание;
- `204 No Content` — успешное удаление;
- `400 Bad Request` — некорректный JSON или неверный `id`;
- `404 Not Found` — элемент не найден.

## Где ставить breakpoint

- В `handleCreate` перед `Decode`.
- В `decodeAndValidateInput`.
- В `create`, `update`, `delete`.
- В `writeJSON`.

## Что логировать при отладке

- метод и путь;
- входной JSON;
- `id` из URL;
- результат валидации;
- статус ответа.
