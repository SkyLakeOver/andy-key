package frontend

// logsJS содержит все функции для работы с журналом (логами)
var logsJS = `
        // ==================== ФУНКЦИИ ДЛЯ ЖУРНАЛА ====================
        // Примечание: основные функции loadLogs() и showLogTab() уже реализованы в frontend_init.go
        // Этот модуль содержит дополнительные функции и инициализацию для журнала
        
        // Функция обновления всех вкладок журнала
        function refreshAllLogs() {
            if (currentLogTab === 'panel') {
                loadLogs('panel');
            } else if (currentLogTab === 'remote') {
                loadLogs('remote');
            } else if (currentLogTab === 'diagnostic') {
                loadLogs('diagnostic');
            }
        }
        
        // Функция очистки отображения журнала (визуально, без удаления данных)
        function clearLogDisplay(logType) {
            var containerId = logType + '-log-container';
            var container = document.getElementById(containerId);
            
            if (!container) return;
            
            var wasEmpty = container.innerHTML.includes('Журнал пуст');
            
            // Сохраняем текущую прокрутку
            var scrollTop = container.scrollTop;
            
            // Очищаем содержимое
            container.innerHTML = '<p class="empty-state">Журнал очищен (данные сохранены в файлах)</p>';
            
            // Восстанавливаем прокрутку
            container.scrollTop = scrollTop;
            
            if (!wasEmpty) {
                showAlert(null, 'Отображение журнала "' + logType + '" очищено', 'success', true);
            }
        }
        
        // Функция экспорта журнала в текстовый файл
        function exportLog(logType) {
            fetch('/api/logs?type=' + logType)
                .then(response => {
                    if (!response.ok) {
                        return response.json().then(data => {
                            throw new Error(data.error || 'Ошибка загрузки журнала');
                        });
                    }
                    return response.json();
                })
                .then(data => {
                    if (!data.entries || data.entries.length === 0) {
                        showAlert(null, 'Журнал пуст, экспорт невозможен', 'warning', true);
                        return;
                    }
                    
                    // Формируем содержимое файла
                    var content = '=== ЖУРНАЛ: ' + logType.toUpperCase() + ' ===\n';
                    content += 'Экспортировано: ' + new Date().toISOString() + '\n';
                    content += 'Количество записей: ' + data.entries.length + '\n';
                    content += '========================================\n\n';
                    
                    data.entries.forEach(entry => {
                        content += entry + '\n';
                    });
                    
                    // Создаем файл для скачивания
                    var blob = new Blob([content], { type: 'text/plain;charset=utf-8' });
                    var url = URL.createObjectURL(blob);
                    var a = document.createElement('a');
                    a.href = url;
                    a.download = 'orchestrator-log-' + logType + '-' + new Date().toISOString().replace(/[:.]/g, '-') + '.txt';
                    document.body.appendChild(a);
                    a.click();
                    setTimeout(() => {
                        document.body.removeChild(a);
                        URL.revokeObjectURL(url);
                    }, 0);
                    
                    showAlert(null, 'Журнал "' + logType + '" успешно экспортирован', 'success', true);
                })
                .catch(error => {
                    showAlert(null, 'Ошибка экспорта журнала: ' + error.message, 'danger', true);
                });
        }
        
        // ==================== ИНИЦИАЛИЗАЦИЯ ОБРАБОТЧИКОВ СОБЫТИЙ ДЛЯ ЖУРНАЛА ====================
        
        document.addEventListener('DOMContentLoaded', function() {
            // Добавляем обработчики для кнопок обновления в каждой вкладке журнала
            var refreshButtons = document.querySelectorAll('#logs-tab .btn[onclick*="loadLogs"]');
            refreshButtons.forEach(button => {
                button.addEventListener('click', function() {
                    var logType = this.getAttribute('onclick').match(/loadLogs\('([^']+)'\)/)[1];
                    loadLogs(logType);
                });
            });
            
            // Добавляем кнопки экспорта в каждую вкладку журнала
            ['panel', 'remote', 'diagnostic'].forEach(logType => {
                var header = document.querySelector('#' + logType + '-log .card-header');
                if (header && !header.querySelector('.export-btn')) {
                    var exportBtn = document.createElement('button');
                    exportBtn.className = 'btn btn-sm export-btn';
                    exportBtn.style.marginLeft = '10px';
                    exportBtn.textContent = 'Экспорт';
                    exportBtn.onclick = function() { exportLog(logType); };
                    header.appendChild(exportBtn);
                }
            });
        });
`