package frontend

// scriptsJS содержит основную логику инициализации, обработки кликов и взаимодействия с сервером
var scriptsJS = `
// ==================== ГЛОБАЛЬНЫЕ ПЕРЕМЕННЫЕ ====================
var addresses = [];
var employees = [];
var workstations = [];
var hosts = [];
var networkEquipment = [];
var networkMFPs = [];
var ipPhones = [];
var credentials = [];
var scripts = [];
var tasks = [];

// DOM элементы
var addressStreet, addressBuilding, addressCabinet, addressCorridor, addressFloor, addressServiceRoom;
var addressTypeCabinet, addressTypeCorridor, addressTypeService;
var addressCabinetGroup, addressCorridorGroup, addressServiceGroup;
var addressAlert, saveAddressBtn, addressesTable;

var employeeFullName, employeeShortName, employeePhoneCity, employeePhoneInternal, employeeAddress;
var employeeAlert, saveEmployeeBtn, employeesTable;

var workstationAddress, workstationVacant, workstationEmployee, workstationInventory, workstationSerial;
var workstationSeals, workstationMonitors, workstationReplacement, workstationReplacementDate, workstationReplacementLetter;
var workstationEmployeeGroup, workstationReplacementGroup;
var workstationAlert, saveWorkstationBtn, workstationsTable;
var workstationsSearch; // Поле поиска для АРМ

var hostAddress, hostEmployee, hostIP, hostSSHPort, hostEnabled;
var hostAlert, saveHostBtn, hostsTable;

var networkEquipmentAddress, networkEquipmentCategory, networkEquipmentType, networkEquipmentModel, networkEquipmentPorts;
var networkEquipmentAlert, saveNetworkEquipmentBtn, networkEquipmentTable;

var networkMFPAddress, networkMFPModel, networkMFPIP, networkMFPHostname, networkMFPSerial;
var networkMFPAlert, saveNetworkMFPBtn, networkMFPsTable;

var ipPhoneAddress, ipPhoneEmployee, ipPhoneIP;
var ipPhoneAlert, saveIPPhoneBtn, ipPhonesTable;

var credUsername, credPassword, credentialAlert, saveCredentialBtn, credentialsTable;
var scriptName, scriptPath, scriptParameters, scriptDescription, scriptAlert, saveScriptBtn, scriptsTable;
var taskName, taskDescription, taskCredential, taskAlert, saveTaskBtn, tasksTable;
var bashCredential, bashHost, bashConsole, bashOutput, bashAlert, bashExecuteBtn;
var bashCommandsContainer, addBashCommandBtn;

var panelLogContainer, remoteLogContainer, diagnosticLogContainer;

// Глобальные переменные для сортировки АРМ
var workstationSortColumn = '';
var workstationSortDirection = 'asc';

// ==================== ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ ====================
// Функция экранирования HTML для предотвращения XSS
function escapeHtml(text) {
	if (!text) return '';
	return text.toString()
		.replace(/&/g, "&amp;")
		.replace(/</g, "&lt;")
		.replace(/>/g, "&gt;")
		.replace(/"/g, "&quot;")
		.replace(/'/g, "&#039;");
}

// Функция отображения алерта
function showAlert(alertElement, message, type, autoHide) {
	if (!alertElement) {
		console.error('Alert element not found');
		return;
	}
	alertElement.textContent = message;
	alertElement.className = 'alert alert-' + type;
	alertElement.style.display = 'block';
	if (autoHide) {
		setTimeout(function() {
			alertElement.style.display = 'none';
		}, 3000);
	}
}

// Функция переключения вкладок
function switchTab(tabName) {
	// Сбрасываем состояние предыдущей вкладки справочника перед переключением
	var currentActive = document.querySelector('.nav-link.active[data-tab]');
	if (currentActive) {
		var currentTabId = currentActive.getAttribute('data-tab');
		if (currentTabId && tables[currentTabId]) {
			tables[currentTabId].resetStateOnly();
		}
	}

	// Скрываем все вкладки
	document.querySelectorAll('.tab-content').forEach(function(tab) {
		tab.classList.remove('active');
	});
	
	// Показываем выбранную вкладку
	var tabElement = document.getElementById(tabName + '-tab');
	if (tabElement) {
		tabElement.classList.add('active');
	}
	
	// Обновляем заголовок страницы
	var titleElement = document.getElementById('page-title');
	if (titleElement) {
		var tabLink = document.querySelector('a[data-tab="' + tabName + '"]');
		if (tabLink) {
			titleElement.textContent = tabLink.textContent;
		}
	}
	
	// Активируем кнопку в сайдбаре
	document.querySelectorAll('.nav-link').forEach(function(link) {
		link.classList.remove('active');
	});
	document.querySelector('a[data-tab="' + tabName + '"]').classList.add('active');
	
	// Загружаем данные при переключении на определённые вкладки
	loadTabData(tabName);
}

// Функция загрузки данных при переключении вкладки
function loadTabData(tabName) {
	switch(tabName) {
		case 'addresses':
			fetch('/api/reference/addresses')
				.then(r => r.json())
				.then(data => {
					addresses = data;
					renderAddressesTable();
					populateAddressDropdowns();
					populateWorkstationDropdowns();
				});
			break;
		case 'employees':
			fetch('/api/reference/addresses')
				.then(r => r.json())
				.then(addrData => {
					addresses = addrData;
					populateAddressDropdowns();
					populateWorkstationDropdowns();
					fetch('/api/reference/employees')
						.then(r => r.json())
						.then(empData => {
							employees = empData;
							renderEmployeesTable();
							populateEmployeeDropdowns();
							populateWorkstationDropdowns();
						});
				});
			break;
		case 'workstations':
			// ИСПРАВЛЕНО: загружаем все необходимые данные для вкладки АРМ
			fetch('/api/reference/addresses')
				.then(r => r.json())
				.then(addrData => {
					addresses = addrData;
					populateAddressDropdowns();
					populateWorkstationDropdowns();
					fetch('/api/reference/employees')
						.then(r => r.json())
						.then(empData => {
							employees = empData;
							populateEmployeeDropdowns();
							populateWorkstationDropdowns();
							fetch('/api/reference/workstations')
								.then(r => r.json())
								.then(wsData => {
									workstations = wsData;
									renderWorkstationsTable();
								});
						});
				});
			break;
		case 'hosts':
			fetch('/api/reference/addresses')
				.then(r => r.json())
				.then(addrData => {
					addresses = addrData;
					populateAddressDropdowns();
					fetch('/api/reference/employees')
						.then(r => r.json())
						.then(empData => {
							employees = empData;
							populateEmployeeDropdowns();
							fetch('/api/reference/hosts')
								.then(r => r.json())
								.then(hostData => {
									hosts = hostData;
									renderHostsTable();
								});
						});
				});
			break;
		case 'network-equipment':
			fetch('/api/reference/addresses')
				.then(r => r.json())
				.then(addrData => {
					addresses = addrData;
					populateAddressDropdowns();
					fetch('/api/reference/network-equipment')
						.then(r => r.json())
						.then(eqData => {
							networkEquipment = eqData;
							renderNetworkEquipmentTable();
						});
				});
			break;
		case 'network-mfps':
			fetch('/api/reference/addresses')
				.then(r => r.json())
				.then(addrData => {
					addresses = addrData;
					populateAddressDropdowns();
					fetch('/api/reference/network-mfps')
						.then(r => r.json())
						.then(mfpData => {
							networkMFPs = mfpData;
							renderNetworkMFPsTable();
						});
				});
			break;
		case 'ip-phones':
			fetch('/api/reference/addresses')
				.then(r => r.json())
				.then(addrData => {
					addresses = addrData;
					populateAddressDropdowns();
					fetch('/api/reference/ip-phones')
						.then(r => r.json())
						.then(phoneData => {
							ipPhones = phoneData;
							renderIPPhonesTable();
						});
				});
			break;
		case 'credentials':
			fetch('/api/credentials')
				.then(r => r.json())
				.then(data => {
					credentials = data;
					renderCredentialsTable();
				});
			break;
		case 'scripts':
			fetch('/api/scripts')
				.then(r => r.json())
				.then(data => {
					scripts = data;
					renderScriptsTable();
				});
			break;
		case 'tasks':
			fetch('/api/tasks')
				.then(r => r.json())
				.then(data => {
					tasks = data;
					renderTasksTable();
				});
			break;
		case 'bash-constructor':
			fetch('/api/credentials')
				.then(r => r.json())
				.then(credData => {
					credentials = credData;
					populateBashDropdowns();
					fetch('/api/reference/hosts')
						.then(r => r.json())
						.then(hostData => {
							hosts = hostData;
							populateBashDropdowns();
						});
				});
			break;
		case 'logs':
			// Логи загружаются при клике на вкладки внутри
			break;
	}
}

// Функция переключения типа адреса
function toggleAddressType() {
	var type = document.querySelector('input[name="address-type"]:checked').value;
	addressCabinetGroup.style.display = type === 'cabinet' ? 'block' : 'none';
	addressCorridorGroup.style.display = type === 'corridor' ? 'block' : 'none';
	addressServiceGroup.style.display = type === 'service' ? 'block' : 'none';
}

// Функция переключения поля сотрудника в АРМ
function toggleWorkstationEmployee() {
	workstationEmployeeGroup.style.display = workstationVacant.checked ? 'none' : 'block';
}

// Функция переключения поля замены оборудования в АРМ
function toggleWorkstationReplacement() {
	workstationReplacementGroup.style.display = workstationReplacement.checked ? 'block' : 'none';
}

// ==================== ФУНКЦИИ ЗАГРУЗКИ ЛОГОВ ====================
function showLogTab(logType) {
	// Скрываем все вкладки логов
	document.querySelectorAll('#logs-tab .tab-content').forEach(function(tab) {
		tab.classList.remove('active');
	});
	
	// Скрываем все кнопки вкладок логов
	document.querySelectorAll('#logs-tab .tab').forEach(function(tab) {
		tab.classList.remove('active');
	});
	
	// Показываем выбранную вкладку логов
	document.getElementById(logType + '-log').classList.add('active');
	document.querySelector('#logs-tab .tab:nth-child(' + 
		(logType === 'panel' ? 1 : logType === 'remote' ? 2 : 3) + 
	')').classList.add('active');
	
	// Загружаем логи
	loadLogs(logType);
}

function loadLogs(logType) {
	fetch('/api/logs?type=' + logType)
		.then(r => r.json())
		.then(data => {
			var container = document.getElementById(logType + '-log-container');
			if (!container) return;
			
			if (data.entries && data.entries.length > 0) {
				container.innerHTML = data.entries.map(function(entry) {
					return '<div>' + escapeHtml(entry) + '</div>';
				}).join('');
			} else {
				container.innerHTML = '<p class="empty-state">Журнал пуст</p>';
			}
		})
		.catch(error => {
			console.error('Ошибка загрузки логов:', error);
		});
}

// ==================== ФУНКЦИИ ДЛЯ BASH-КОНСТРУКТОРА ====================
function populateBashDropdowns() {
	// Заполняем селект учётных записей
	if (bashCredential) {
		bashCredential.innerHTML = '<option value="">-- Выберите учётную запись --</option>';
		credentials.forEach(function(cred) {
			var option = document.createElement('option');
			option.value = cred.id;
			option.textContent = cred.username;
			bashCredential.appendChild(option);
		});
	}
	
	// Заполняем селект хостов
	if (bashHost) {
		bashHost.innerHTML = '<option value="">-- Выберите хост --</option>';
		hosts.forEach(function(host) {
			var option = document.createElement('option');
			option.value = host.id;
			option.textContent = host.full_address + ' (' + host.ip + ')';
			bashHost.appendChild(option);
		});
	}
}

function executeBashCommand() {
	var credentialId = parseInt(bashCredential.value);
	var hostId = parseInt(bashHost.value);
	var command = bashConsole.value.trim();
	
	if (!credentialId) {
		showAlert(bashAlert, 'Выберите учётную запись для подключения', 'danger');
		return;
	}
	
	if (!hostId) {
		showAlert(bashAlert, 'Выберите хост для выполнения команды', 'danger');
		return;
	}
	
	if (!command) {
		showAlert(bashAlert, 'Введите команду для выполнения', 'danger');
		return;
	}
	
	// Очищаем вывод
	bashOutput.innerHTML = '<div class="info">Выполнение команды...</div>';
	
	// Отправляем запрос на сервер
	fetch('/api/execute-command', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({
			host_id: hostId,
			credential_id: credentialId,
			command: command
		})
	})
	.then(response => {
		if (!response.ok) {
			return response.json().then(data => {
				throw new Error(data.error || 'Ошибка сервера');
			});
		}
		return response.json();
	})
	.then(result => {
		// Отображаем результат
		if (result.status === 'completed') {
			bashOutput.innerHTML = '<div class="success">Команда выполнена успешно на ' + escapeHtml(result.host) + ':</div><pre>' + escapeHtml(result.output) + '</pre>';
		} else {
			bashOutput.innerHTML = '<div class="error">Ошибка выполнения команды:</div><pre>' + escapeHtml(result.output || result.error) + '</pre>';
		}
	})
	.catch(error => {
		bashOutput.innerHTML = '<div class="error">Ошибка: ' + escapeHtml(error.message) + '</div>';
		showAlert(bashAlert, error.message || 'Ошибка сети', 'danger');
	});
}

function addBashCommand() {
	var row = document.createElement('div');
	row.className = 'bash-command-row';
	row.style.display = 'flex';
	row.style.gap = '10px';
	row.style.marginBottom = '10px';
	
	var input = document.createElement('input');
	input.type = 'text';
	input.className = 'form-control bash-command-input';
	input.placeholder = 'Введите команду bash';
	
	var button = document.createElement('button');
	button.className = 'btn btn-sm btn-danger';
	button.textContent = 'Удалить';
	button.onclick = function() {
		this.parentElement.remove();
	};
	
	row.appendChild(input);
	row.appendChild(button);
	bashCommandsContainer.insertBefore(row, addBashCommandBtn);
}

function removeBashCommand(button) {
	button.parentElement.remove();
}

// ==================== ФУНКЦИИ СОРТИРОВКИ И ПОИСКА ДЛЯ АРМ ====================
// Функция сортировки таблицы АРМ
function sortWorkstationsTable(column) {
	if (workstationSortColumn === column) {
		workstationSortDirection = workstationSortDirection === 'asc' ? 'desc' : 'asc';
	} else {
		workstationSortColumn = column;
		workstationSortDirection = 'asc';
	}
	
	// Обновляем заголовки
	document.querySelectorAll('#workstations-table th').forEach(function(th) {
		th.classList.remove('asc', 'desc');
	});
	
	var header = document.querySelector('#workstations-table th[onclick="sortWorkstationsTable(\'' + column + '\')"]');
	if (header) {
		header.classList.add(workstationSortDirection);
	}
	
	// Сортируем данные
	workstations.sort(function(a, b) {
		var valA, valB;
		
		if (column === 'replacement') {
			// Сортировка по полю замены оборудования
			valA = a.replacement_done === '1' || a.replacement_done === 1 || a.replacement_done === true;
			valB = b.replacement_done === '1' || b.replacement_done === 1 || b.replacement_done === true;
		} else if (column === 'is_vacant') {
			// Сортировка по статусу вакантности
			valA = a.is_vacant === '1' || a.is_vacant === 1 || a.is_vacant === true;
			valB = b.is_vacant === '1' || b.is_vacant === 1 || b.is_vacant === true;
		} else {
			// Сортировка по другим полям
			valA = a[column] || '';
			valB = b[column] || '';
		}
		
		if (typeof valA === 'string' && typeof valB === 'string') {
			valA = valA.toLowerCase();
			valB = valB.toLowerCase();
		}
		
		if (valA < valB) return workstationSortDirection === 'asc' ? -1 : 1;
		if (valA > valB) return workstationSortDirection === 'asc' ? 1 : -1;
		return 0;
	});
	
	renderWorkstationsTable();
}

// Функция поиска в таблице АРМ
function searchWorkstations() {
	var query = workstationsSearch.value.toLowerCase().trim();
	
	if (!query) {
		// Показываем все записи
		workstations.forEach(function(ws) {
			delete ws._hidden;
		});
	} else {
		// Фильтруем по подстроке
		workstations.forEach(function(ws) {
			var match = false;
			
			// Проверяем все поля
			if (ws.full_address && ws.full_address.toLowerCase().includes(query)) match = true;
			if (ws.employee_short_name && ws.employee_short_name.toLowerCase().includes(query)) match = true;
			if (ws.inventory_number && ws.inventory_number.toLowerCase().includes(query)) match = true;
			if (ws.serial_number && ws.serial_number.toLowerCase().includes(query)) match = true;
			if (ws.monitor_count && ws.monitor_count.toString().includes(query)) match = true;
			if (ws.replacement_date && ws.replacement_date.toLowerCase().includes(query)) match = true;
			if (ws.replacement_letter && ws.replacement_letter.toLowerCase().includes(query)) match = true;
			
			ws._hidden = !match;
		});
	}
	
	renderWorkstationsTable();
}

// ==================== ФУНКЦИЯ ВЫХОДА ====================
function logout() {
	fetch('/logout', { method: 'POST' })
		.then(() => {
			window.location.href = '/login';
		})
		.catch(error => {
			console.error('Ошибка выхода:', error);
			window.location.href = '/login';
		});
}

// ==================== ИНИЦИАЛИЗАЦИЯ ПРИ ЗАГРУЗКЕ СТРАНИЦЫ ====================
document.addEventListener('DOMContentLoaded', function() {
	// Инициализация переменных (получение ссылок на элементы)
	// Адреса
	addressStreet = document.getElementById('address-street');
	addressBuilding = document.getElementById('address-building');
	addressCabinet = document.getElementById('address-cabinet');
	addressCorridor = document.getElementById('address-corridor');
	addressFloor = document.getElementById('address-floor');
	addressServiceRoom = document.getElementById('address-service-room');
	addressTypeCabinet = document.getElementById('address-type-cabinet');
	addressTypeCorridor = document.getElementById('address-type-corridor');
	addressTypeService = document.getElementById('address-type-service');
	addressCabinetGroup = document.getElementById('address-cabinet-group');
	addressCorridorGroup = document.getElementById('address-corridor-group');
	addressServiceGroup = document.getElementById('address-service-group');
	addressAlert = document.getElementById('address-alert');
	saveAddressBtn = document.getElementById('save-address');
	addressesTable = document.getElementById('addresses-table').getElementsByTagName('tbody')[0];
	
	// Сотрудники
	employeeFullName = document.getElementById('employee-full-name');
	employeeShortName = document.getElementById('employee-short-name');
	employeePhoneCity = document.getElementById('employee-phone-city');
	employeePhoneInternal = document.getElementById('employee-phone-internal');
	employeeAddress = document.getElementById('employee-address');
	employeeAlert = document.getElementById('employee-alert');
	saveEmployeeBtn = document.getElementById('save-employee');
	employeesTable = document.getElementById('employees-table').getElementsByTagName('tbody')[0];
	
	// АРМ
	workstationAddress = document.getElementById('workstation-address');
	workstationVacant = document.getElementById('workstation-vacant');
	workstationEmployee = document.getElementById('workstation-employee');
	workstationInventory = document.getElementById('workstation-inventory');
	workstationSerial = document.getElementById('workstation-serial');
	workstationSeals = document.getElementById('workstation-seals');
	workstationMonitors = document.getElementById('workstation-monitors');
	workstationReplacement = document.getElementById('workstation-replacement');
	workstationReplacementDate = document.getElementById('workstation-replacement-date');
	workstationReplacementLetter = document.getElementById('workstation-replacement-letter');
	workstationEmployeeGroup = document.getElementById('workstation-employee-group');
	workstationReplacementGroup = document.getElementById('workstation-replacement-group');
	workstationAlert = document.getElementById('workstation-alert');
	saveWorkstationBtn = document.getElementById('save-workstation');
	workstationsTable = document.getElementById('workstations-table');
	workstationsSearch = document.getElementById('workstations-search'); // Поле поиска
	
	// Хосты
	hostAddress = document.getElementById('host-address');
	hostEmployee = document.getElementById('host-employee');
	hostIP = document.getElementById('host-ip');
	hostSSHPort = document.getElementById('host-ssh-port');
	hostEnabled = document.getElementById('host-enabled');
	hostAlert = document.getElementById('host-alert');
	saveHostBtn = document.getElementById('save-host');
	hostsTable = document.getElementById('hosts-table');
	
	// Сетевое оборудование
	networkEquipmentAddress = document.getElementById('network-equipment-address');
	networkEquipmentCategory = document.getElementById('network-equipment-category');
	networkEquipmentType = document.getElementById('network-equipment-type');
	networkEquipmentModel = document.getElementById('network-equipment-model');
	networkEquipmentPorts = document.getElementById('network-equipment-ports');
	networkEquipmentAlert = document.getElementById('network-equipment-alert');
	saveNetworkEquipmentBtn = document.getElementById('save-network-equipment');
	networkEquipmentTable = document.getElementById('network-equipment-table');
	
	// Сетевые МФУ
	networkMFPAddress = document.getElementById('network-mfp-address');
	networkMFPModel = document.getElementById('network-mfp-model');
	networkMFPIP = document.getElementById('network-mfp-ip');
	networkMFPHostname = document.getElementById('network-mfp-hostname');
	networkMFPSerial = document.getElementById('network-mfp-serial');
	networkMFPAlert = document.getElementById('network-mfp-alert');
	saveNetworkMFPBtn = document.getElementById('save-network-mfp');
	networkMFPsTable = document.getElementById('network-mfps-table');
	
	// IP телефоны
	ipPhoneAddress = document.getElementById('ip-phone-address');
	ipPhoneEmployee = document.getElementById('ip-phone-employee');
	ipPhoneIP = document.getElementById('ip-phone-ip');
	ipPhoneAlert = document.getElementById('ip-phone-alert');
	saveIPPhoneBtn = document.getElementById('save-ip-phone');
	ipPhonesTable = document.getElementById('ip-phones-table');
	
	// Учётные записи
	credUsername = document.getElementById('cred-username');
	credPassword = document.getElementById('cred-password');
	credentialAlert = document.getElementById('credential-alert');
	saveCredentialBtn = document.getElementById('save-credential');
	credentialsTable = document.getElementById('credentials-table');
	
	// Скрипты
	scriptName = document.getElementById('script-name');
	scriptPath = document.getElementById('script-path');
	scriptParameters = document.getElementById('script-parameters');
	scriptDescription = document.getElementById('script-description');
	scriptAlert = document.getElementById('script-alert');
	saveScriptBtn = document.getElementById('save-script');
	scriptsTable = document.getElementById('scripts-table');
	
	// Задачи
	taskName = document.getElementById('task-name');
	taskDescription = document.getElementById('task-description');
	taskCredential = document.getElementById('task-credential');
	taskAlert = document.getElementById('task-alert');
	saveTaskBtn = document.getElementById('save-task');
	tasksTable = document.getElementById('tasks-table');
	
	// BASH-конструктор
	bashCredential = document.getElementById('bash-credential');
	bashHost = document.getElementById('bash-host');
	bashConsole = document.getElementById('bash-console');
	bashOutput = document.getElementById('bash-output');
	bashAlert = document.getElementById('bash-alert');
	bashExecuteBtn = document.getElementById('bash-execute-btn');
	bashCommandsContainer = document.getElementById('bash-commands-container');
	addBashCommandBtn = document.getElementById('add-bash-command-btn');
	
	// Логи
	panelLogContainer = document.getElementById('panel-log-container');
	remoteLogContainer = document.getElementById('remote-log-container');
	diagnosticLogContainer = document.getElementById('diagnostic-log-container');
	
	// Обработчики событий для переключения типа адреса
	if (addressTypeCabinet) addressTypeCabinet.addEventListener('change', toggleAddressType);
	if (addressTypeCorridor) addressTypeCorridor.addEventListener('change', toggleAddressType);
	if (addressTypeService) addressTypeService.addEventListener('change', toggleAddressType);
	
	// Обработчики событий для переключения поля сотрудника в АРМ
	if (workstationVacant) workstationVacant.addEventListener('change', toggleWorkstationEmployee);
	
	// Обработчики событий для переключения поля замены оборудования в АРМ
	if (workstationReplacement) workstationReplacement.addEventListener('change', toggleWorkstationReplacement);
	
	// Обработчики событий для BASH-конструктора
	if (bashExecuteBtn) bashExecuteBtn.addEventListener('click', executeBashCommand);
	if (addBashCommandBtn) addBashCommandBtn.addEventListener('click', addBashCommand);
	
	// Обработчики событий для поиска в АРМ
	if (workstationsSearch) {
		workstationsSearch.addEventListener('input', searchWorkstations);
	}
	
	// ИСПРАВЛЕНО: загружаем все необходимые данные при старте
	// Сначала адреса
	fetch('/api/reference/addresses')
		.then(r => r.json())
		.then(addrData => {
			addresses = addrData;
			populateAddressDropdowns();
			populateWorkstationDropdowns();
			// Затем сотрудники
			fetch('/api/reference/employees')
				.then(r => r.json())
				.then(empData => {
					employees = empData;
					populateEmployeeDropdowns();
					populateWorkstationDropdowns();
					// Затем АРМ (для вкладки, если она будет открыта)
					// Загружаем остальные данные по мере переключения вкладок
				});
		});
	
	// Загружаем учётные записи для BASH-конструктора
	fetch('/api/credentials')
		.then(r => r.json())
		.then(credData => {
			credentials = credData;
		});
	
	// Загружаем хосты для BASH-конструктора
	// (будут загружены при переключении на вкладку)
	
	// Устанавливаем текущего пользователя (если есть)
	// (это можно сделать через куки или через запрос к серверу)
});
`