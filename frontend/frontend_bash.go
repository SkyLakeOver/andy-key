package frontend

// bashJS содержит все функции для BASH-конструктора
var bashJS = `
        // ==================== ФУНКЦИИ ДЛЯ BASH-КОНСТРУКТОРА ====================
        
        // Загрузка учётных записей в BASH-конструктор
        function loadBashCredentials() {
            var select = document.getElementById('bash-credential');
            if (!select) return;
            
            select.innerHTML = '<option value="">-- Выберите учётную запись --</option>';
            
            credentials.forEach(cred => {
                var option = document.createElement('option');
                option.value = cred.id;
                option.textContent = cred.username;
                select.appendChild(option);
            });
            
            // Инициализация обработчика кнопки выполнения (только один раз)
            var bashExecuteBtn = document.getElementById('bash-execute-btn');
            if (bashExecuteBtn && !bashExecuteBtn.initialized) {
                bashExecuteBtn.addEventListener('click', executeBashCommand);
                bashExecuteBtn.initialized = true;
            }
        }
        
        // Загрузка хостов в BASH-конструктор
        function loadBashHosts() {
            var select = document.getElementById('bash-host');
            if (!select) return;
            
            select.innerHTML = '<option value="">-- Выберите хост --</option>';
            
            // Сортируем хосты по адресу
            var sortedHosts = [...hosts].sort((a, b) => {
                if (a.full_address !== b.full_address) return a.full_address.localeCompare(b.full_address);
                return (a.employee_short_name || '').localeCompare(b.employee_short_name || '');
            });
            
            sortedHosts.forEach(host => {
                var option = document.createElement('option');
                option.value = host.id;
                var displayText = host.full_address + 
                                 (host.employee_short_name ? ' | ' + host.employee_short_name : '') + 
                                 ' | ' + host.ip;
                option.textContent = displayText;
                select.appendChild(option);
            });
        }
        
        // Выполнение команды в BASH-конструкторе
        function executeBashCommand() {
            var credentialId = parseInt(document.getElementById('bash-credential').value);
            var hostId = parseInt(document.getElementById('bash-host').value);
            var command = bashConsole.value.trim();
            
            if (!credentialId) {
                showAlert(bashAlert, 'Выберите учётную запись для подключения', 'danger');
                return;
            }
            
            if (!hostId) {
                showAlert(bashAlert, 'Выберите хост', 'danger');
                return;
            }
            
            if (!command) {
                showAlert(bashAlert, 'Введите команду для выполнения', 'danger');
                return;
            }
            
            // Очищаем вывод и показываем статус выполнения
            bashOutput.innerHTML = '<div class="info">Выполнение команды на хосте... Пожалуйста, подождите.</div>';
            if (bashAlert) bashAlert.style.display = 'none';
            
            var data = {
                host_id: hostId,
                credential_id: credentialId,
                command: command
            };
            
            fetch('/api/execute-command', {
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
                // Форматируем вывод для отображения
                var outputHtml = '';
                if (result.status === 'completed') {
                    outputHtml += '<div class="success">[УСПЕХ] Команда выполнена успешно</div>\n\n';
                    outputHtml += '<div class="info">Результат выполнения:</div>\n';
                    outputHtml += '<pre style="color:#f8f8f2;white-space:pre-wrap;background-color:#2d2d2d;padding:10px;border-radius:4px;">' + 
                        result.output.replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/\n/g, '<br>') + 
                        '</pre>';
                } else {
                    outputHtml += '<div class="error">[ОШИБКА] ' + (result.error || 'Неизвестная ошибка') + '</div>';
                }
                bashOutput.innerHTML = outputHtml;
                
                // Логируем успешное выполнение
                if (result.status === 'completed') {
                    showAlert(null, 'Команда выполнена успешно', 'success', true);
                }
            })
            .catch(error => {
                bashOutput.innerHTML = '<div class="error">[ОШИБКА] ' + error.message + '</div>';
                showAlert(bashAlert, error.message || 'Ошибка выполнения команды', 'danger');
            });
        }
`