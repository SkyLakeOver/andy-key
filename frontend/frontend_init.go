package frontend

// initJS содержит инициализацию и вспомогательные функции
var initJS = `
        // ==================== ИНИЦИАЛИЗАЦИЯ И ГЛОБАЛЬНЫЕ ПЕРЕМЕННЫЕ ====================
        
        // DOM Elements
        var tabs = document.querySelectorAll('.nav-link');
        var tabContents = document.querySelectorAll('.tab-content');
        var pageTitle = document.getElementById('page-title');
        
        // State
        var credentials = [];
        var scripts = [];
        var tasks = [];
        var addresses = [];
        var employees = [];
        var workstations = [];
        var hosts = [];
        var networkEquipment = [];
        var networkMFPs = [];
        var ipPhones = [];
        var currentLogTab = 'panel';
        
        // Form elements for Tasks
        var taskName = document.getElementById('task-name');
        var taskDescription = document.getElementById('task-description');
        var taskAlert = document.getElementById('task-alert');
        var saveTaskBtn = document.getElementById('save-task');
        var addBashCommandBtn = document.getElementById('add-bash-command-btn');
        var addTaskScriptBtn = document.getElementById('add-task-script-btn');
        var taskCredential = document.getElementById('task-credential');
        var taskHostsList = document.getElementById('task-hosts-list');
        var bashCommandsContainer = document.getElementById('bash-commands-container');
        var taskScriptsContainer = document.getElementById('task-scripts-container');
        
        // Form elements for Credentials
        var credUsername = document.getElementById('cred-username');
        var credPassword = document.getElementById('cred-password');
        var credentialAlert = document.getElementById('credential-alert');
        var saveCredentialBtn = document.getElementById('save-credential');
        
        // Form elements for Scripts
        var scriptName = document.getElementById('script-name');
        var scriptPath = document.getElementById('script-path');
        var scriptParameters = document.getElementById('script-parameters');
        var scriptDescription = document.getElementById('script-description');
        var scriptAlert = document.getElementById('script-alert');
        var saveScriptBtn = document.getElementById('save-script');
        
        // Form elements for BASH-constructor
        var bashCredential = document.getElementById('bash-credential');
        var bashHost = document.getElementById('bash-host');
        var bashConsole = document.getElementById('bash-console');
        var bashExecuteBtn = document.getElementById('bash-execute-btn');
        var bashOutput = document.getElementById('bash-output');
        var bashAlert = document.getElementById('bash-alert');
        
        // Form elements for Addresses
        var addressStreet = document.getElementById('address-street');
        var addressBuilding = document.getElementById('address-building');
        var addressCabinet = document.getElementById('address-cabinet');
        var addressCorridor = document.getElementById('address-corridor');
        var addressFloor = document.getElementById('address-floor');
        var addressServiceRoom = document.getElementById('address-service-room');
        var addressAlert = document.getElementById('address-alert');
        var saveAddressBtn = document.getElementById('save-address');
        var addressTypeCabinet = document.getElementById('address-type-cabinet');
        var addressTypeCorridor = document.getElementById('address-type-corridor');
        var addressTypeService = document.getElementById('address-type-service');
        var addressCabinetGroup = document.getElementById('address-cabinet-group');
        var addressCorridorGroup = document.getElementById('address-corridor-group');
        var addressServiceGroup = document.getElementById('address-service-group');
        
        // Form elements for Employees
        var employeeFullName = document.getElementById('employee-full-name');
        var employeeShortName = document.getElementById('employee-short-name');
        var employeePhoneCity = document.getElementById('employee-phone-city');
        var employeePhoneInternal = document.getElementById('employee-phone-internal');
        var employeeAddress = document.getElementById('employee-address');
        var employeeAlert = document.getElementById('employee-alert');
        var saveEmployeeBtn = document.getElementById('save-employee');
        
        // Form elements for Workstations
        var workstationAddress = document.getElementById('workstation-address');
        var workstationEmployee = document.getElementById('workstation-employee');
        var workstationVacant = document.getElementById('workstation-vacant');
        var workstationInventory = document.getElementById('workstation-inventory');
        var workstationSerial = document.getElementById('workstation-serial');
        var workstationSeals = document.getElementById('workstation-seals');
        var workstationMonitors = document.getElementById('workstation-monitors');
        var workstationReplacement = document.getElementById('workstation-replacement');
        var workstationReplacementDate = document.getElementById('workstation-replacement-date');
        var workstationReplacementLetter = document.getElementById('workstation-replacement-letter');
        var workstationAlert = document.getElementById('workstation-alert');
        var saveWorkstationBtn = document.getElementById('save-workstation');
        var workstationEmployeeGroup = document.getElementById('workstation-employee-group');
        var workstationReplacementGroup = document.getElementById('workstation-replacement-group');
        
        // Form elements for Hosts
        var hostAddress = document.getElementById('host-address');
        var hostEmployee = document.getElementById('host-employee');
        var hostIP = document.getElementById('host-ip');
        var hostSSHPort = document.getElementById('host-ssh-port');
        var hostEnabled = document.getElementById('host-enabled');
        var hostAlert = document.getElementById('host-alert');
        var saveHostBtn = document.getElementById('save-host');
        
        // Form elements for Network Equipment
        var networkEquipmentAddress = document.getElementById('network-equipment-address');
        var networkEquipmentCategory = document.getElementById('network-equipment-category');
        var networkEquipmentType = document.getElementById('network-equipment-type');
        var networkEquipmentModel = document.getElementById('network-equipment-model');
        var networkEquipmentPorts = document.getElementById('network-equipment-ports');
        var networkEquipmentAlert = document.getElementById('network-equipment-alert');
        var saveNetworkEquipmentBtn = document.getElementById('save-network-equipment');
        
        // Form elements for Network MFPs
        var networkMFPAddress = document.getElementById('network-mfp-address');
        var networkMFPModel = document.getElementById('network-mfp-model');
        var networkMFPIP = document.getElementById('network-mfp-ip');
        var networkMFPHostname = document.getElementById('network-mfp-hostname');
        var networkMFPSerial = document.getElementById('network-mfp-serial');
        var networkMFPAlert = document.getElementById('network-mfp-alert');
        var saveNetworkMFPBtn = document.getElementById('save-network-mfp');
        
        // Form elements for IP Phones
        var ipPhoneAddress = document.getElementById('ip-phone-address');
        var ipPhoneEmployee = document.getElementById('ip-phone-employee');
        var ipPhoneIP = document.getElementById('ip-phone-ip');
        var ipPhoneAlert = document.getElementById('ip-phone-alert');
        var saveIPPhoneBtn = document.getElementById('save-ip-phone');
        
        // Tables
        var credentialsTable = document.getElementById('credentials-table')?.getElementsByTagName('tbody')[0];
        var scriptsTable = document.getElementById('scripts-table')?.getElementsByTagName('tbody')[0];
        var tasksTable = document.getElementById('tasks-table')?.getElementsByTagName('tbody')[0];
        var addressesTable = document.getElementById('addresses-table')?.getElementsByTagName('tbody')[0];
        var employeesTable = document.getElementById('employees-table')?.getElementsByTagName('tbody')[0];
        var workstationsTable = document.getElementById('workstations-table')?.getElementsByTagName('tbody')[0];
        var hostsTable = document.getElementById('hosts-table')?.getElementsByTagName('tbody')[0];
        var networkEquipmentTable = document.getElementById('network-equipment-table')?.getElementsByTagName('tbody')[0];
        var networkMFPsTable = document.getElementById('network-mfps-table')?.getElementsByTagName('tbody')[0];
        var ipPhonesTable = document.getElementById('ip-phones-table')?.getElementsByTagName('tbody')[0];
        
        // ==================== ИНИЦИАЛИЗАЦИЯ ПРИ ЗАГРУЗКЕ ====================
        
        document.addEventListener('DOMContentLoaded', function() {
            // Загружаем данные для аутентификации
            fetch('/api/credentials')
                .then(response => {
                    if (response.status === 401) {
                        window.location.href = '/login';
                        return;
                    }
                    return response.json();
                })
                .then(data => {
                    credentials = data;
                    document.getElementById('current-user').textContent = 'Администратор';
                    
                    // Загружаем все данные справочников
                    Promise.all([
                        fetch('/api/reference/addresses').then(r => r.json()),
                        fetch('/api/reference/employees').then(r => r.json()),
                        fetch('/api/reference/workstations').then(r => r.json()),
                        fetch('/api/reference/hosts').then(r => r.json()),
                        fetch('/api/reference/network-equipment').then(r => r.json()),
                        fetch('/api/reference/network-mfps').then(r => r.json()),
                        fetch('/api/reference/ip-phones').then(r => r.json()),
                        fetch('/api/scripts').then(r => r.json()),
                        fetch('/api/tasks').then(r => r.json())
                    ])
                    .then(([addr, emp, ws, hst, ne, nm, ip, scr, tsk]) => {
                        addresses = addr;
                        employees = emp;
                        workstations = ws;
                        hosts = hst;
                        networkEquipment = ne;
                        networkMFPs = nm;
                        ipPhones = ip;
                        scripts = scr;
                        tasks = tsk;
                        
                        // Рендерим все таблицы
                        if (typeof renderAddressesTable === 'function') renderAddressesTable();
                        if (typeof renderEmployeesTable === 'function') renderEmployeesTable();
                        if (typeof renderWorkstationsTable === 'function') renderWorkstationsTable();
                        if (typeof renderHostsTable === 'function') renderHostsTable();
                        if (typeof renderNetworkEquipmentTable === 'function') renderNetworkEquipmentTable();
                        if (typeof renderNetworkMFPsTable === 'function') renderNetworkMFPsTable();
                        if (typeof renderIPPhonesTable === 'function') renderIPPhonesTable();
                        if (typeof renderScriptsTable === 'function') renderScriptsTable();
                        if (typeof renderTasksTable === 'function') renderTasksTable();
                        
                        // Заполняем выпадающие списки
                        if (typeof populateAddressDropdowns === 'function') populateAddressDropdowns();
                        if (typeof populateEmployeeDropdowns === 'function') populateEmployeeDropdowns();
                        if (typeof loadTaskCredentials === 'function') loadTaskCredentials();
                        if (typeof loadTaskHostsList === 'function') loadTaskHostsList();
                        if (typeof loadBashCredentials === 'function') loadBashCredentials();
                        if (typeof loadBashHosts === 'function') loadBashHosts();
                    })
                    .catch(error => {
                        console.error('Ошибка загрузки данных:', error);
                    });
                    
                    // Инициализируем обработчики событий
                    if (typeof setupEventListeners === 'function') setupEventListeners();
                    
                    // Загружаем журнал при открытии вкладки
                    document.querySelectorAll('.nav-link').forEach(tab => {
                        tab.addEventListener('click', function(e) {
                            e.preventDefault();
                            var targetTab = this.getAttribute('data-tab');
                            if (typeof switchTab === 'function') switchTab(targetTab);
                            
                            if (targetTab === 'logs' && currentLogTab) {
                                if (typeof loadLogs === 'function') loadLogs(currentLogTab);
                            }
                        });
                    });
                })
                .catch(error => {
                    if (error.message.includes('401')) {
                        window.location.href = '/login';
                    }
                });
        });
        
        // ==================== НАВИГАЦИЯ ====================
        
        function switchTab(tabName) {
            // Скрываем все вкладки
            tabs.forEach(tab => tab.classList.remove('active'));
            tabContents.forEach(content => content.classList.remove('active'));
            
            // Активируем выбранную вкладку
            var targetTabLink = document.querySelector('.nav-link[data-tab="' + tabName + '"]');
            if (targetTabLink) targetTabLink.classList.add('active');
            
            var targetContent = document.getElementById(tabName + '-tab');
            if (targetContent) targetContent.classList.add('active');
            
            // Обновляем заголовок страницы
            var titles = {
                'tasks': 'Задачи',
                'credentials': 'Учётные записи SSH',
                'scripts': 'Скрипты',
                'bash-constructor': 'BASH-конструктор',
                'logs': 'Журнал',
                'addresses': 'Адреса',
                'employees': 'Сотрудники',
                'workstations': 'АРМ',
                'hosts': 'Хосты',
                'network-equipment': 'Сетевое оборудование',
                'network-mfps': 'Сетевые МФУ',
                'ip-phones': 'IP телефоны'
            };
            pageTitle.textContent = titles[tabName] || 'SSH Orchestrator';
        }
        
        // ==================== ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ ====================
        
        // Функция экранирования HTML
        function escapeHtml(unsafe) {
            if (!unsafe) return '';
            return unsafe
                .replace(/&/g, "&amp;")
                .replace(/</g, "&lt;")
                .replace(/>/g, "&gt;")
                .replace(/"/g, "&quot;")
                .replace(/'/g, "&#039;");
        }
        
        // Функция отображения алертов
        function showAlert(element, message, type, global) {
            if (global) {
                var alertDiv = document.createElement('div');
                alertDiv.className = 'alert alert-' + type;
                alertDiv.textContent = message;
                alertDiv.style.position = 'fixed';
                alertDiv.style.top = '20px';
                alertDiv.style.right = '20px';
                alertDiv.style.zIndex = '3000';
                alertDiv.style.minWidth = '300px';
                document.body.appendChild(alertDiv);
                
                setTimeout(() => {
                    alertDiv.style.opacity = '0';
                    alertDiv.style.transition = 'opacity 0.5s';
                    setTimeout(() => document.body.removeChild(alertDiv), 500);
                }, 3000);
                return;
            }
            
            if (element) {
                element.textContent = message;
                element.className = 'alert alert-' + type;
                element.style.display = 'block';
                
                // Скрываем через 5 секунд
                setTimeout(() => {
                    element.style.display = 'none';
                }, 5000);
            }
        }
        
        // Функция заполнения выпадающих списков адресов
        function populateAddressDropdowns() {
            // Сортируем адреса для выпадающих списков
            var sortedAddresses = [...addresses].sort((a, b) => {
                if (a.street !== b.street) return a.street.localeCompare(b.street);
                if (a.building !== b.building) return a.building.localeCompare(b.building);
                if (a.floor !== b.floor) return (a.floor || 0) - (b.floor || 0);
                if (a.cabinet !== b.cabinet) return (a.cabinet || '').localeCompare(b.cabinet || '');
                if (a.corridor !== b.corridor) return (a.corridor || '').localeCompare(b.corridor || '');
                return (a.service_room || '').localeCompare(b.service_room || '');
            });
            
            // Заполняем все выпадающие списки адресов
            var addressDropdowns = [
                employeeAddress,
                workstationAddress,
                hostAddress,
                networkEquipmentAddress,
                networkMFPAddress,
                ipPhoneAddress
            ];
            
            addressDropdowns.forEach(dropdown => {
                if (!dropdown) return;
                dropdown.innerHTML = '<option value="">-- Выберите адрес --</option>';
                sortedAddresses.forEach(addr => {
                    var option = document.createElement('option');
                    option.value = addr.id;
                    option.textContent = addr.full_address;
                    dropdown.appendChild(option);
                });
            });
        }
        
        // Функция заполнения выпадающих списков сотрудников
        function populateEmployeeDropdowns() {
            // Сортируем сотрудников по ФИО
            var sortedEmployees = [...employees].sort((a, b) => a.full_name.localeCompare(b.full_name));
            
            // Заполняем выпадающие списки сотрудников
            var employeeDropdowns = [
                workstationEmployee,
                hostEmployee
            ];
            
            employeeDropdowns.forEach(dropdown => {
                if (!dropdown) return;
                dropdown.innerHTML = '<option value="">-- Не закреплён за сотрудником --</option>';
                sortedEmployees.forEach(emp => {
                    var option = document.createElement('option');
                    option.value = emp.id;
                    option.textContent = emp.short_name || emp.full_name;
                    dropdown.appendChild(option);
                });
            });
        }
        
        // Функции переключения типов адресов
        function toggleAddressType() {
            if (addressTypeCabinet.checked) {
                addressCabinetGroup.style.display = 'block';
                addressCorridorGroup.style.display = 'none';
                addressServiceGroup.style.display = 'none';
                addressCabinet.required = true;
                addressCorridor.required = false;
                addressFloor.required = false;
                addressServiceRoom.required = false;
            } else if (addressTypeCorridor.checked) {
                addressCabinetGroup.style.display = 'none';
                addressCorridorGroup.style.display = 'block';
                addressServiceGroup.style.display = 'none';
                addressCabinet.required = false;
                addressCorridor.required = true;
                addressFloor.required = true;
                addressServiceRoom.required = false;
            } else if (addressTypeService.checked) {
                addressCabinetGroup.style.display = 'none';
                addressCorridorGroup.style.display = 'none';
                addressServiceGroup.style.display = 'block';
                addressCabinet.required = false;
                addressCorridor.required = false;
                addressFloor.required = false;
                addressServiceRoom.required = true;
            }
        }
        
        // Функции переключения вакантного места АРМ
        function toggleWorkstationEmployee() {
            if (workstationVacant.checked) {
                workstationEmployeeGroup.style.display = 'none';
                workstationEmployee.required = false;
            } else {
                workstationEmployeeGroup.style.display = 'block';
                workstationEmployee.required = true;
            }
        }
        
        // Функции переключения замены оборудования АРМ
        function toggleWorkstationReplacement() {
            if (workstationReplacement.checked) {
                workstationReplacementGroup.style.display = 'block';
                workstationReplacementDate.required = true;
            } else {
                workstationReplacementGroup.style.display = 'none';
                workstationReplacementDate.required = false;
            }
        }
        
        // ==================== ФУНКЦИИ ВЫХОДА И ЗАГРУЗКИ ЖУРНАЛА ====================
        
        function logout() {
            fetch('/logout', { method: 'POST' })
                .then(() => {
                    window.location.href = '/login';
                });
        }
        
        function loadLogs(logType) {
            var containerId = logType + '-log-container';
            var container = document.getElementById(containerId);
            
            if (!container) return;
            
            container.innerHTML = '<p class="empty-state">Загрузка...</p>';
            
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
                    if (data.entries && data.entries.length > 0) {
                        var html = '';
                        data.entries.forEach(entry => {
                            html += '<div style="padding:4px 0;border-bottom:1px solid #f0f0f0;font-family:monospace;font-size:0.85rem;">' + entry + '</div>';
                        });
                        container.innerHTML = html;
                    } else {
                        container.innerHTML = '<p class="empty-state">Журнал пуст</p>';
                    }
                })
                .catch(error => {
                    console.error('Ошибка загрузки журнала:', error);
                    container.innerHTML = '<p class="empty-state">Ошибка загрузки журнала: ' + error.message + '</p>';
                });
        }
        
        function showLogTab(tabName) {
            // Скрываем все вкладки журнала
            var logTabs = document.querySelectorAll('#logs-tab .tab');
            var logContents = document.querySelectorAll('#logs-tab .tab-content');
            
            logTabs.forEach(tab => tab.classList.remove('active'));
            logContents.forEach(content => content.classList.remove('active'));
            
            // Активируем выбранную вкладку
            event.target.classList.add('active');
            document.getElementById(tabName + '-log').classList.add('active');
            currentLogTab = tabName;
            
            // Загружаем логи
            loadLogs(tabName);
        }
        
        // ==================== ИНИЦИАЛИЗАЦИЯ ОБРАБОТЧИКОВ СОБЫТИЙ ====================
        
        function setupEventListeners() {
            // Tab Navigation
            tabs.forEach(tab => {
                tab.addEventListener('click', function(e) {
                    e.preventDefault();
                    var targetTab = this.getAttribute('data-tab');
                    switchTab(targetTab);
                });
            });
            
            // Form Submissions - будут добавлены в других модулях
            // Здесь только базовая инициализация
            
            // Address type radio buttons
            if (addressTypeCabinet) addressTypeCabinet.addEventListener('change', toggleAddressType);
            if (addressTypeCorridor) addressTypeCorridor.addEventListener('change', toggleAddressType);
            if (addressTypeService) addressTypeService.addEventListener('change', toggleAddressType);
            
            // Workstation vacant checkbox
            if (workstationVacant) workstationVacant.addEventListener('change', toggleWorkstationEmployee);
            
            // Workstation replacement checkbox
            if (workstationReplacement) workstationReplacement.addEventListener('change', toggleWorkstationReplacement);
        }
`