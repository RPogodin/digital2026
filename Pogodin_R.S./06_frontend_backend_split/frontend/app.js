const taskList = document.getElementById("task-list");
const taskForm = document.getElementById("task-form");
const titleInput = document.getElementById("title");
const statusNode = document.getElementById("status");
const reloadButton = document.getElementById("reload-button");

// loadTasks показывает основной сценарий frontend: GET -> JSON -> обновление DOM.
async function loadTasks() {
    setStatus("Загружаю список задач...");

    try {
        const response = await fetch("/api/tasks");
        const tasks = await parseJSON(response);

        if (!response.ok) {
            throw new Error(tasks.error || "Не удалось получить список задач");
        }

        renderTasks(tasks);
        setStatus("Список задач обновлен");
    } catch (error) {
        setStatus(error.message, true);
    }
}

// createTask собирает данные формы и отправляет их на backend через fetch.
async function createTask(event) {
    event.preventDefault();

    const title = titleInput.value.trim();
    if (!title) {
        setStatus("Введите заголовок задачи", true);
        titleInput.focus();
        return;
    }

    setStatus("Отправляю новую задачу на backend...");

    try {
        const response = await fetch("/api/tasks", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({ title }),
        });

        const payload = await parseJSON(response);
        if (!response.ok) {
            throw new Error(payload.error || "Не удалось создать задачу");
        }

        taskForm.reset();
        await loadTasks();
        setStatus(`Задача создана: ${payload.title}`);
    } catch (error) {
        setStatus(error.message, true);
    }
}

// deleteTask отправляет DELETE и потом заново загружает список.
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

        await loadTasks();
        setStatus(`Задача #${taskId} удалена`);
    } catch (error) {
        setStatus(error.message, true);
    }
}

// renderTasks отвечает только за обновление DOM на основе уже готовых данных.
function renderTasks(tasks) {
    taskList.innerHTML = "";

    if (tasks.length === 0) {
        const emptyNode = document.createElement("li");
        emptyNode.className = "empty";
        emptyNode.textContent = "Пока нет задач. Создайте первую запись через форму выше.";
        taskList.appendChild(emptyNode);
        return;
    }

    for (const task of tasks) {
        const item = document.createElement("li");
        item.className = "task-card";

        const createdAt = new Date(task.created_at).toLocaleString("ru-RU");

        item.innerHTML = `
            <div>
                <p class="task-title">${escapeHTML(task.title)}</p>
                <p class="task-meta">Создано: ${escapeHTML(createdAt)} | ID: ${task.id}</p>
            </div>
            <button type="button" data-delete-id="${task.id}">Удалить</button>
        `;

        taskList.appendChild(item);
    }
}

function setStatus(message, isError = false) {
    statusNode.textContent = message;
    statusNode.style.color = isError ? "#9c2f1b" : "";
}

async function parseJSON(response) {
    const text = await response.text();
    return text ? JSON.parse(text) : {};
}

function escapeHTML(value) {
    return value
        .replaceAll("&", "&amp;")
        .replaceAll("<", "&lt;")
        .replaceAll(">", "&gt;")
        .replaceAll('"', "&quot;");
}

taskForm.addEventListener("submit", createTask);
reloadButton.addEventListener("click", loadTasks);

// Делегирование событий позволяет не навешивать обработчик на каждую кнопку отдельно.
taskList.addEventListener("click", (event) => {
    const button = event.target.closest("button[data-delete-id]");
    if (!button) {
        return;
    }

    const taskId = Number(button.dataset.deleteId);
    deleteTask(taskId);
});

loadTasks();
