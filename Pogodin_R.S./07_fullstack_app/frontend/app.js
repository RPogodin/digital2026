const taskForm = document.getElementById("task-form");
const formTitle = document.getElementById("form-title");
const taskIdInput = document.getElementById("task-id");
const titleInput = document.getElementById("title");
const descriptionInput = document.getElementById("description");
const statusInput = document.getElementById("status");
const categoryInput = document.getElementById("category");
const submitButton = document.getElementById("submit-button");
const resetButton = document.getElementById("reset-button");
const reloadButton = document.getElementById("reload-button");
const filterButton = document.getElementById("filter-button");
const searchInput = document.getElementById("search");
const filterStatusInput = document.getElementById("filter-status");
const filterCategoryInput = document.getElementById("filter-category");
const taskList = document.getElementById("task-list");
const statusMessage = document.getElementById("status-message");

let tasks = [];

// loadTasks читает фильтры из формы, делает GET /api/tasks и потом перерисовывает список.
async function loadTasks() {
    setStatus("Загружаю задачи...");

    const params = new URLSearchParams();
    const query = searchInput.value.trim();
    const status = filterStatusInput.value;
    const category = filterCategoryInput.value;

    if (query) {
        params.set("q", query);
    }
    if (status) {
        params.set("status", status);
    }
    if (category) {
        params.set("category", category);
    }

    const url = params.toString() ? `/api/tasks?${params.toString()}` : "/api/tasks";

    try {
        const response = await fetch(url);
        const payload = await parseJSON(response);

        if (!response.ok) {
            throw new Error(payload.error || "Не удалось загрузить задачи");
        }

        tasks = payload;
        renderTasks(tasks);
        setStatus(`Загружено задач: ${tasks.length}`);
    } catch (error) {
        setStatus(error.message, true);
    }
}

// saveTask обслуживает и создание, и редактирование, чтобы было видно сходство POST и PUT.
async function saveTask(event) {
    event.preventDefault();

    const payload = {
        title: titleInput.value.trim(),
        description: descriptionInput.value.trim(),
        status: statusInput.value,
        category: categoryInput.value,
    };

    if (!payload.title) {
        setStatus("Поле заголовка обязательно", true);
        titleInput.focus();
        return;
    }

    const taskId = taskIdInput.value;
    const isEdit = Boolean(taskId);
    const url = isEdit ? `/api/tasks/${taskId}` : "/api/tasks";
    const method = isEdit ? "PUT" : "POST";

    setStatus(isEdit ? `Обновляю задачу #${taskId}...` : "Создаю задачу...");

    try {
        const response = await fetch(url, {
            method,
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify(payload),
        });

        const result = await parseJSON(response);
        if (!response.ok) {
            throw new Error(result.error || "Не удалось сохранить задачу");
        }

        resetForm();
        await loadTasks();
        setStatus(isEdit ? `Задача #${result.id} обновлена` : `Задача #${result.id} создана`);
    } catch (error) {
        setStatus(error.message, true);
    }
}

// deleteTask показывает короткий сценарий DELETE -> обновление списка.
async function deleteTask(taskId) {
    setStatus(`Удаляю задачу #${taskId}...`);

    try {
        const response = await fetch(`/api/tasks/${taskId}`, {
            method: "DELETE",
        });

        if (!response.ok) {
            const payload = await parseJSON(response);
            throw new Error(payload.error || "Не удалось удалить задачу");
        }

        if (taskIdInput.value === String(taskId)) {
            resetForm();
        }

        await loadTasks();
        setStatus(`Задача #${taskId} удалена`);
    } catch (error) {
        setStatus(error.message, true);
    }
}

// editTask не идет на сервер, а только заполняет форму данными выбранной задачи.
function editTask(taskId) {
    const task = tasks.find((item) => item.id === taskId);
    if (!task) {
        setStatus(`Не удалось найти задачу #${taskId} в текущем списке`, true);
        return;
    }

    taskIdInput.value = task.id;
    titleInput.value = task.title;
    descriptionInput.value = task.description;
    statusInput.value = task.status;
    categoryInput.value = task.category;
    submitButton.textContent = "Сохранить изменения";
    formTitle.textContent = `Редактирование задачи #${task.id}`;
    setStatus(`Форма заполнена данными задачи #${task.id}`);
    titleInput.focus();
}

// resetForm возвращает страницу в режим создания новой задачи.
function resetForm() {
    taskForm.reset();
    taskIdInput.value = "";
    statusInput.value = "todo";
    categoryInput.value = "lecture";
    submitButton.textContent = "Создать задачу";
    formTitle.textContent = "Новая задача";
}

// renderTasks обновляет DOM на основе последнего ответа backend.
function renderTasks(list) {
    taskList.innerHTML = "";

    if (list.length === 0) {
        const emptyNode = document.createElement("li");
        emptyNode.className = "empty";
        emptyNode.textContent = "По текущему фильтру задач не найдено.";
        taskList.appendChild(emptyNode);
        return;
    }

    for (const task of list) {
        const item = document.createElement("li");
        item.className = "task-card";
        item.innerHTML = `
            <h3>${escapeHTML(task.title)}</h3>
            <div class="task-meta">
                <span class="chip status-${escapeHTML(task.status)}">${escapeHTML(task.status)}</span>
                <span class="chip">${escapeHTML(task.category)}</span>
                <span class="chip">Создано: ${escapeHTML(formatDate(task.created_at))}</span>
                <span class="chip">ID: ${task.id}</span>
            </div>
            <p>${escapeHTML(task.description || "Описание не заполнено")}</p>
            <div class="task-actions">
                <button type="button" data-edit-id="${task.id}" class="ghost-button">Редактировать</button>
                <button type="button" data-delete-id="${task.id}">Удалить</button>
            </div>
        `;

        taskList.appendChild(item);
    }
}

function setStatus(message, isError = false) {
    statusMessage.textContent = message;
    statusMessage.style.color = isError ? "#9f2f17" : "";
}

function formatDate(value) {
    return new Date(value).toLocaleString("ru-RU");
}

async function parseJSON(response) {
    const text = await response.text();
    return text ? JSON.parse(text) : {};
}

function escapeHTML(value) {
    return String(value)
        .replaceAll("&", "&amp;")
        .replaceAll("<", "&lt;")
        .replaceAll(">", "&gt;")
        .replaceAll('"', "&quot;");
}

taskForm.addEventListener("submit", saveTask);
resetButton.addEventListener("click", () => {
    resetForm();
    setStatus("Форма очищена");
});
reloadButton.addEventListener("click", loadTasks);
filterButton.addEventListener("click", loadTasks);

// Один обработчик на список проще объяснять, чем много отдельных подписок на кнопки.
taskList.addEventListener("click", (event) => {
    const editButton = event.target.closest("button[data-edit-id]");
    if (editButton) {
        editTask(Number(editButton.dataset.editId));
        return;
    }

    const deleteButton = event.target.closest("button[data-delete-id]");
    if (deleteButton) {
        deleteTask(Number(deleteButton.dataset.deleteId));
    }
});

searchInput.addEventListener("keydown", (event) => {
    if (event.key === "Enter") {
        event.preventDefault();
        loadTasks();
    }
});

resetForm();
loadTasks();
