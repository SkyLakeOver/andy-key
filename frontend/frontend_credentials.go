package frontend

// credentialsJS содержит все функции для работы с учётными записями SSH
var credentialsJS = `
        // ==================== ФУНКЦИИ ДЛЯ УЧЁТНЫХ ЗАПИСЕЙ SSH ====================
        
        function loadCredentials() {
            fetch('/api/credentials')
                .then(response => {
                    if (!response.ok) {
                        return response.json().then(data => {
                            throw new Error(data.error || 'Ошибка загрузки учётных записей');
                        });
                    }
                    return response.json();
                })
                .then(data => {
                    credentials = data;
                    renderCredentialsTable();
                    loadTaskCredentials();
                    loadBashCredentials();
                })
                .catch(error => {
                    console.error('Error loading credentials:', error);
                    if (credentialsTable) {
                        credentialsTable.innerHTML = '<tr><td colspan="3" class="empty-state">Ошибка загрузки учётных записей: ' + error.message + '</td></tr>';
                    }
                });
        }
        
        function renderCredentialsTable() {
            if (!credentialsTable) return;
            
            if (credentials.length === 0) {
                credentialsTable.innerHTML = '<tr><td colspan="3" class="empty-state">Учётные записи не найдены</td></tr>';
                return;
            }
            
            var html = '';
            credentials.forEach(cred => {
                html += '<tr data-id="' + cred.id + '">' +
                    '<td>' + cred.id + '</td>' +
                    '<td><strong>' + cred.username + '</strong></td>' +
                    '<td>' +
                        '<button class="btn btn-sm btn-edit edit-btn" onclick="editCredentialRow(' + cred.id + ', this)">✏️</button>' +
                        '<button class="btn btn-sm btn-save edit-btn" onclick="saveCredentialRow(' + cred.id + ', this)" style="display:none;">💾</button>' +
                        '<button class="btn btn-sm btn-cancel edit-btn" onclick="cancelCredentialEdit(' + cred.id + ', this)" style="display:none;">❌</button>' +
                        '<button class="btn btn-sm btn-danger" onclick="deleteCredential(' + cred.id + ')">Удалить</button>' +
                    '</td>' +
                '</tr>';
            });
            credentialsTable.innerHTML = html;
        }
        
        function saveCredential() {
            var username = credUsername.value.trim();
            var password = credPassword.value;
            
            if (!username || !password) {
                showAlert(credentialAlert, 'Логин и пароль обязательны', 'danger');
                return;
            }
            
            if (password.trim() === '') {
                showAlert(credentialAlert, 'Пароль не может состоять только из пробелов', 'danger');
                return;
            }
            
            var bodyData = { username: username, password: password };
            
            fetch('/api/credentials', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(bodyData)
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
                credUsername.value = '';
                credPassword.value = '';
                showAlert(null, 'Учётная запись добавлена!', 'success', true);
                loadCredentials();
            })
            .catch(error => {
                showAlert(credentialAlert, error.message || 'Ошибка сети', 'danger');
            });
        }
        
        function deleteCredential(id) {
            if (confirm('Вы уверены, что хотите удалить эту учётную запись? Это действие нельзя отменить.')) {
                // В реальном приложении здесь будет запрос на удаление
                // Пока просто обновляем список
                loadCredentials();
            }
        }
        
        // Режим редактирования учётной записи
        function editCredentialRow(credId, button) {
            var row = button.closest('tr');
            var usernameCell = row.cells[1];
            
            // Сохраняем оригинальное значение
            var originalUsername = usernameCell.querySelector('strong').textContent;
            row.setAttribute('data-original', originalUsername);
            
            // Преобразуем ячейку в поле ввода
            usernameCell.innerHTML = '<input type="text" class="form-control" value="' + escapeHtml(originalUsername) + '">';
            
            // Показываем кнопки сохранения/отмены
            row.querySelector('.btn-edit').style.display = 'none';
            row.querySelector('.btn-save').style.display = 'inline-block';
            row.querySelector('.btn-cancel').style.display = 'inline-block';
            row.classList.add('edit-mode');
        }
        
        // Сохранение изменений учётной записи
        function saveCredentialRow(credId, button) {
            var row = button.closest('tr');
            var input = row.querySelector('input');
            var newUsername = input.value.trim();
            
            if (!newUsername) {
                showAlert(credentialAlert, 'Логин обязателен', 'danger');
                return;
            }
            
            // Отправляем только логин (пароль не меняем при редактировании в таблице)
            var data = { username: newUsername };
            
            fetch('/api/credentials/' + credId, {
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
                var index = credentials.findIndex(c => c.id == credId);
                if (index !== -1) {
                    credentials[index].username = newUsername;
                }
                loadCredentials();
                showAlert(null, 'Учётная запись успешно обновлена!', 'success', true);
            })
            .catch(error => {
                showAlert(credentialAlert, error.message || 'Ошибка сети', 'danger');
            });
        }
        
        // Отмена редактирования учётной записи
        function cancelCredentialEdit(credId, button) {
            var row = button.closest('tr');
            var originalUsername = row.getAttribute('data-original');
            
            // Восстанавливаем оригинальное значение
            row.cells[1].innerHTML = '<strong>' + escapeHtml(originalUsername) + '</strong>';
            
            // Скрываем кнопки сохранения/отмены
            row.querySelector('.btn-edit').style.display = 'inline-block';
            row.querySelector('.btn-save').style.display = 'none';
            row.querySelector('.btn-cancel').style.display = 'none';
            row.classList.remove('edit-mode');
        }
        
        // ==================== ИНИЦИАЛИЗАЦИЯ ОБРАБОТЧИКОВ СОБЫТИЙ ДЛЯ УЧЁТНЫХ ЗАПИСЕЙ ====================
        
        document.addEventListener('DOMContentLoaded', function() {
            // Credential handlers
            if (saveCredentialBtn) saveCredentialBtn.addEventListener('click', saveCredential);
        });
`