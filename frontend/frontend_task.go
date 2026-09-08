package frontend

// tasksJS содержит все функции для работы с задачами
var tasksJS = `
        // ==================== ФУНКЦИИ ДЛЯ ЗАДАЧ ====================
        
        function loadTasks() {
            fetch('/api/tasks')
                .then(response => {
                    if (!response.ok) {
                        return response.json().then(data => {
                            throw new Error(data.error || 'Ошибка загрузки задач');
                        });
                    }
                    return response.json();
                })
                .then(data => {
                    tasks = data;
                    renderTasksTable();
                })
                .catch(error => {
                    console.error('Error loading tasks:', error);
                    if (tasksTable) {
                        tasksTable.innerHTML = '<tr><td colspan="6" class="empty-state">Ошибка загрузки задач: ' + error.message + '</td></tr>';
                    }
                });
        }
        
        function renderTasksTable() {
            if (!tasksTable) return;
            
            if (tasks.length === 0) {
                tasksTable.innerHTML = '<tr><td colspan="6" class="empty-state">Задачи не найдены</td></tr>';
                return;
            }
            
            var html = '';
            tasks.forEach(task => {
                var statusClass = '';
                if (task.status === 'running') statusClass = 'task-status-running';
                else if (task.status === 'completed') statusClass = 'task-status-completed';
                else if (task.status === 'failed') statusClass = 'task-status-failed';
                else if (task.status === 'canceled') statusClass = 'task-status-canceled';
                else statusClass = 'task-status-pending';
                
                html += '<tr data-id="' + task.id + '">' +
                    '<td>' + task.id + '</td>' +
                    '<td><strong>' + task.name + '</strong></td>' +
                    '<td>' + (task.description || '-') + '</td>' +
                    '<td class="' + statusClass + '">' + task.status + '</td>' +
                    '<td>' + new Date(task.created_at).toLocaleDateString() + '</td>' +
                    '<td>' +
                        '<button class="btn btn-sm btn-edit edit-btn" onclick="editTaskRow(' + task.id + ', this)">✏️</button>' +
                        '<button class="btn btn-sm btn-save edit-btn" onclick="saveTaskRow(' + task.id + ', this)" style="display:none;">💾</button>' +
                        '<button class="btn btn-sm btn-cancel edit-btn" onclick="cancelTaskEdit(' + task.id + ', this)" style="display:none;">❌</button>' +
                        '<button class="btn btn-sm btn-success" onclick="startTask(' + task.id + ')">Старт</button> ' +
                        '<button class="btn btn-sm btn-danger" onclick="stopTask(' + task.id + ')">Стоп</button> ' +
                        '<button class="btn btn-sm btn-secondary" onclick="deleteTask(' + task.id + ')">Удалить</button>' +
                    '</td>' +
                '</tr>';
            });
            tasksTable.innerHTML = html;
        }
        
        function loadTaskCredentials() {
            var select = document.getElementById('task-credential');
            if (!select) return;
            
            select.innerHTML = '<option value="">-- Выберите учётную запись --</option>';
            
            credentials.forEach(cred => {
                var option = document.createElement('option');
                option.value = cred.id;
                option.textContent = cred.username;
                select.appendChild(option);
            });
        }
        
        function loadTaskHostsList() {
            var container = document.getElementById('task-hosts-list');
            if (!container) return;
            
            if (hosts.length === 0) {
                container.innerHTML = '<p class="empty-state">Сначала добавьте хосты в разделе "Справочники → Хосты"</p>';
                return;
            }
            
            // Сортировка: улица → дом → этаж → кабинет/коридор
            var sortedHosts = [...hosts].sort((a, b) => {
                if (a.full_address !== b.full_address) return a.full_address.localeCompare(b.full_address);
                return (a.employee_short_name || '').localeCompare(b.employee_short_name || '');
            });
            
            var html = '';
            sortedHosts.forEach(host => {
                // Формат: "Улица, д.Дом, каб.Кабинет | ФИО | IP-адрес"
                var displayText = host.full_address + 
                                  (host.employee_short_name ? ' | ' + host.employee_short_name : '') + 
                                  ' | ' + host.ip;
                
                html += '<div class="host-checkbox-item" style="padding:8px;border-bottom:1px solid #eee;display:flex;align-items:center;gap:10px;">' +
                    '<input type="checkbox" class="task-host-checkbox" value="' + host.id + '" id="host-' + host.id + '">' +
                    '<label for="host-' + host.id + '" style="margin:0;">' + displayText + '</label>' +
                '</div>';
            });
            
            container.innerHTML = html;
        }
        
        function addBashCommandRow() {
            var container = document.getElementById('bash-commands-container');
            var rowIndex = container.children.length;
            
            var div = document.createElement('div');
            div.className = 'bash-command-row';
            div.style.display = 'flex';
            div.style.gap = '10px';
            div.style.marginBottom = '10px';
            div.innerHTML = '<input type="text" class="form-control bash-command-input" placeholder="Введите команду bash (например: apt update)">' +
                '<button class="btn btn-sm btn-danger" onclick="removeBashCommand(this)">Удалить</button>';
            
            container.appendChild(div);
        }
        
        function removeBashCommand(button) {
            var row = button.parentElement;
            row.parentElement.removeChild(row);
            
            // Если не осталось команд, добавляем одну пустую
            var container = document.getElementById('bash-commands-container');
            if (container.children.length === 0) {
                addBashCommandRow();
            }
        }
        
        function addTaskScriptRow() {
            if (scripts.length === 0) {
                showAlert(taskAlert, 'Сначала добавьте скрипты в библиотеку', 'danger');
                return;
            }
            
            var rowIndex = document.querySelectorAll('.task-script-row').length;
            
            var html = '<div class="task-script-row" id="task-script-row-' + rowIndex + '" style="background:#f8f9fa;padding:15px;margin-bottom:15px;border-radius:4px;border:1px solid #ddd;display:flex;align-items:center;gap:10px;">' +
                '<select class="task-script-select" style="flex:2;">' +
                    '<option value="">-- Выберите скрипт --</option>';
            
            scripts.forEach(script => {
                html += '<option value="' + script.id + '">' + script.name + '</option>';
            });
            
            html += '</select>' +
                '<textarea class="task-script-parameters" placeholder=\'{"param1": "value1"}\' style="flex:2;padding:8px;border:1px solid #ddd;border-radius:4px;margin-left:10px;" rows="1"></textarea>' +
                '<div class="task-controls" style="display:flex;gap:5px;">' +
                    '<button class="btn btn-sm btn-secondary" onclick="moveTaskScriptUp(' + rowIndex + ')">↑</button>' +
                    '<button class="btn btn-sm btn-secondary" onclick="moveTaskScriptDown(' + rowIndex + ')">↓</button>' +
                    '<button class="btn btn-sm btn-danger" onclick="removeTaskScriptRow(' + rowIndex + ')">Удалить</button>' +
                '</div>' +
            '</div>';
            
            if (taskScriptsContainer.querySelector('.empty-state')) {
                taskScriptsContainer.innerHTML = '';
            }
            taskScriptsContainer.insertAdjacentHTML('beforeend', html);
        }
        
        function removeTaskScriptRow(index) {
            var row = document.getElementById('task-script-row-' + index);
            if (row) row.remove();
            if (taskScriptsContainer.children.length === 0) {
                taskScriptsContainer.innerHTML = '<p class="empty-state">Добавьте скрипты для задачи</p>';
            }
        }
        
        function moveTaskScriptUp(index) {
            var row = document.getElementById('task-script-row-' + index);
            var prev = row.previousElementSibling;
            if (prev) {
                row.parentNode.insertBefore(row, prev);
            }
        }
        
        function moveTaskScriptDown(index) {
            var row = document.getElementById('task-script-row-' + index);
            var next = row.nextElementSibling;
            if (next) {
                row.parentNode.insertBefore(next, row);
            }
        }
        
        function saveTask() {
            var name = taskName.value.trim();
            var description = taskDescription.value.trim();
            var credentialId = parseInt(document.getElementById('task-credential').value);
            
            if (!name) {
                showAlert(taskAlert, 'Название задачи обязательно', 'danger');
                return;
            }
            
            if (!credentialId) {
                showAlert(taskAlert, 'Выберите учётную запись для задачи', 'danger');
                return;
            }
            
            // Сбор отмеченных хостов
            var hostCheckboxes = document.querySelectorAll('.task-host-checkbox:checked');
            var hostIds = [];
            hostCheckboxes.forEach(checkbox => {
                hostIds.push(parseInt(checkbox.value));
            });
            
            if (hostIds.length === 0) {
                showAlert(taskAlert, 'Выберите хотя бы один хост для задачи', 'danger');
                return;
            }
            
            // Сбор команд bash
            var bashInputs = document.querySelectorAll('.bash-command-input');
            var bashCommands = [];
            bashInputs.forEach(input => {
                var cmd = input.value.trim();
                if (cmd !== '') {
                    bashCommands.push(cmd);
                }
            });
            
            // Сбор скриптов
            var taskScriptRows = document.querySelectorAll('.task-script-row');
            var taskScripts = [];
            
            for (var i = 0; i < taskScriptRows.length; i++) {
                var scriptSelect = taskScriptRows[i].querySelector('.task-script-select');
                var paramsInput = taskScriptRows[i].querySelector('.task-script-parameters');
                var scriptId = parseInt(scriptSelect.value);
                var parameters = paramsInput.value.trim() || '{}';
                
                if (!scriptId) {
                    showAlert(taskAlert, 'Выберите скрипт для шага ' + (i + 1), 'danger');
                    return;
                }
                
                try {
                    JSON.parse(parameters);
                } catch (e) {
                    showAlert(taskAlert, 'Некорректный JSON в параметрах скрипта ' + (i + 1) + ': ' + e.message, 'danger');
                    return;
                }
                
                taskScripts.push({
                    script_id: scriptId,
                    parameters: parameters,
                    order_index: i
                });
            }
            
            if (bashCommands.length === 0 && taskScripts.length === 0) {
                showAlert(taskAlert, 'Добавьте хотя бы одну команду bash или скрипт', 'danger');
                return;
            }
            
            var data = {
                name: name,
                description: description,
                credential_id: credentialId,
                host_ids: hostIds,
                bash_commands: bashCommands,
                scripts: taskScripts
            };
            
            fetch('/api/tasks', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(data)
            })
            .then(response => {
                if (!response.ok) {
                    return response.json().then(data => {
                        throw new Error(data.error || 'Неизвестная ошибка сервера');
                    });
                }
                return response.json();
            })
            .then(result => {
                taskName.value = '';
                taskDescription.value = '';
                document.getElementById('task-credential').value = '';
                document.querySelectorAll('.task-host-checkbox').forEach(cb => cb.checked = false);
                document.getElementById('bash-commands-container').innerHTML = '';
                addBashCommandRow();
                document.getElementById('task-scripts-container').innerHTML = '<p class="empty-state">Добавьте скрипты для задачи</p>';
                
                showAlert(null, 'Задача "' + name + '" создана!', 'success', true);
                loadTasks();
            })
            .catch(error => {
                showAlert(taskAlert, error.message || 'Ошибка сети', 'danger');
            });
        }
        
        // Режим редактирования задачи
        function editTaskRow(taskId, button) {
            var row = button.closest('tr');
            var cells = row.querySelectorAll('td');
            
            // Сохраняем оригинальные значения
            var originalValues = [];
            for (var i = 0; i < cells.length - 1; i++) {
                originalValues.push(cells[i].innerHTML);
            }
            row.setAttribute('data-original', JSON.stringify(originalValues));
            
            // Находим задачу в данных
            var task = tasks.find(t => t.id == taskId);
            if (!task) return;
            
            // Преобразуем ячейки в поля ввода
            cells[1].innerHTML = '<input type="text" class="form-control" value="' + escapeHtml(task.name) + '">';
            cells[2].innerHTML = '<textarea class="form-control" rows="2">' + escapeHtml(task.description || '') + '</textarea>';
            
            // Показываем кнопки сохранения/отмены
            row.querySelector('.btn-edit').style.display = 'none';
            row.querySelector('.btn-save').style.display = 'inline-block';
            row.querySelector('.btn-cancel').style.display = 'inline-block';
            row.classList.add('edit-mode');
        }
        
        // Сохранение изменений задачи
        function saveTaskRow(taskId, button) {
            var row = button.closest('tr');
            var inputs = row.querySelectorAll('input, textarea');
            
            var data = {
                name: inputs[0].value.trim(),
                description: inputs[1].value.trim()
            };
            
            // Валидация
            if (!data.name) {
                showAlert(taskAlert, 'Название задачи обязательно', 'danger');
                return;
            }
            
            fetch('/api/tasks/' + taskId, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(data)
            })
            .then(response => {
                if (!response.ok) {
                    return response.json().then(data => {
                        throw new Error(data.error || 'Неизвестная ошибка сервера');
                    });
                }
                return response.json();
            })
            .then(result => {
                // Обновляем данные в памяти
                var index = tasks.findIndex(t => t.id == taskId);
                if (index !== -1) {
                    tasks[index] = Object.assign(tasks[index], data);
                }
                loadTasks();
                showAlert(null, 'Задача успешно обновлена!', 'success', true);
            })
            .catch(error => {
                showAlert(taskAlert, error.message || 'Ошибка сети', 'danger');
            });
        }
        
        // Отмена редактирования задачи
        function cancelTaskEdit(taskId, button) {
            var row = button.closest('tr');
            var originalValues = JSON.parse(row.getAttribute('data-original'));
            
            // Восстанавливаем оригинальные значения
            var cells = row.querySelectorAll('td');
            for (var i = 0; i < originalValues.length; i++) {
                cells[i].innerHTML = originalValues[i];
            }
            
            // Скрываем кнопки сохранения/отмены
            row.querySelector('.btn-edit').style.display = 'inline-block';
            row.querySelector('.btn-save').style.display = 'none';
            row.querySelector('.btn-cancel').style.display = 'none';
            row.classList.remove('edit-mode');
        }
        
        function startTask(taskId) {
            fetch('/api/tasks/' + taskId + '/control', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ action: 'start' })
            })
            .then(response => {
                if (!response.ok) {
                    return response.json().then(data => {
                        throw new Error(data.error || 'Ошибка запуска задачи');
                    });
                }
                return response.json();
            })
            .then(result => {
                showAlert(null, 'Задача #' + taskId + ' запущена', 'success', true);
                loadTasks();
            })
            .catch(error => {
                showAlert(null, error.message || 'Ошибка сети', 'danger', true);
            });
        }
        
        function stopTask(taskId) {
            if (!confirm('Вы уверены, что хотите остановить эту задачу?')) {
                return;
            }
            
            fetch('/api/tasks/' + taskId + '/control', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ action: 'stop' })
            })
            .then(response => {
                if (!response.ok) {
                    return response.json().then(data => {
                        throw new Error(data.error || 'Ошибка остановки задачи');
                    });
                }
                return response.json();
            })
            .then(result => {
                showAlert(null, 'Задача #' + taskId + ' остановлена', 'success', true);
                loadTasks();
            })
            .catch(error => {
                showAlert(null, error.message || 'Ошибка сети', 'danger', true);
            });
        }
        
        function deleteTask(id) {
            if (confirm('Вы уверены, что хотите удалить эту задачу? Это действие нельзя отменить.')) {
                // В реальном приложении здесь будет запрос на удаление
                // Пока просто обновляем список
                loadTasks();
            }
        }
        
        // ==================== ИНИЦИАЛИЗАЦИЯ ОБРАБОТЧИКОВ СОБЫТИЙ ДЛЯ ЗАДАЧ ====================
        
        document.addEventListener('DOMContentLoaded', function() {
            // Task handlers
            if (saveTaskBtn) saveTaskBtn.addEventListener('click', saveTask);
            if (addBashCommandBtn) addBashCommandBtn.addEventListener('click', addBashCommandRow);
            if (addTaskScriptBtn) addTaskScriptBtn.addEventListener('click', addTaskScriptRow);
        });
`