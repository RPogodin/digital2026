# Этап 06. Раздельные frontend и backend

## Учебная цель

Показать явное разделение:

- backend на Go;
- frontend как статические HTML/CSS/JS файлы;
- связь между ними через `fetch` и JSON API.

Поток данных:

`форма в браузере -> JS -> fetch -> HTTP-запрос -> Go handler -> JSON -> JS -> DOM`

## Как запустить

Из каталога этапа:

```bash
go run .
```

Из корня репозитория:

```bash
make run-06
```

## Какие URL открыть

- `http://localhost:18086/` — frontend
- `http://localhost:18086/api/tasks` — backend API

## Что такое frontend и backend

Frontend — это то, что работает в браузере:

- HTML;
- CSS;
- JavaScript;
- работа с DOM и событиями пользователя.

Backend — это то, что работает на сервере:

- принимает HTTP-запросы;
- выполняет логику;
- отдает JSON.

В этом этапе они запущены одной командой, но логически разделены по каталогам:

- `frontend/` — статические файлы;
- `backend/` — Go-код API.
- `main.go` — точка сборки, которая соединяет эти части в один сервер.

## Как браузер делает fetch-запрос

1. Пользователь вводит заголовок задачи и нажимает кнопку.
2. JavaScript перехватывает отправку формы.
3. JS делает `fetch("/api/tasks", { method: "POST", ... })`.
4. Go backend принимает JSON и создает задачу.
5. Backend возвращает JSON-ответ.
6. JS снова запрашивает список и обновляет DOM.

## Где смотреть запросы в DevTools

В браузере откройте DevTools:

- вкладка `Network` — видно сами HTTP-запросы и ответы;
- вкладка `Console` — удобно проверять ошибки JavaScript;
- вкладка `Elements` — видно, как JS обновляет DOM.

## Что умеет приложение

- загрузить список задач;
- создать задачу;
- удалить задачу.

Backend ожидает один понятный JSON-объект без лишних полей. Если отправить
неверный формат, сервер вернет `400 Bad Request` с текстом ошибки в JSON.

## Примеры curl

Получить список:

```bash
curl -i http://localhost:18086/api/tasks
```

Создать задачу:

```bash
curl -i -X POST http://localhost:18086/api/tasks \
  -H "Content-Type: application/json" \
  -d '{"title":"Новая задача с frontend"}'
```

Удалить задачу:

```bash
curl -i -X DELETE http://localhost:18086/api/tasks/1
```

## Где ставить breakpoint

Backend:

- `backend.handleCreate`
- `backend.create`
- `backend.handleDelete`

Frontend:

- функция `createTask` в `frontend/app.js`
- функция `loadTasks`
- обработчик клика на кнопке удаления

## Что логировать при отладке

- тело JSON, которое отправляет frontend;
- HTTP status ответа;
- список задач после перезагрузки;
- `taskId`, который уходит в `DELETE`.

## Архитектурное решение

Используется один Go-сервер:

- `/api/...` обслуживает backend API;
- `/` и файлы `app.js`, `styles.css` отдаются как статические ресурсы.

Так запуск остается простым, но разделение ролей frontend и backend видно явно.

Дополнительно на этом этапе появился простой HTTP-лог в `main.go`, чтобы во время
лекции можно было видеть не только API-вызовы, но и загрузку самой страницы,
`app.js` и `styles.css`.
