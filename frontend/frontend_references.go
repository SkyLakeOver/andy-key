package frontend

// referencesJS содержит все функции для справочников (адреса, сотрудники, АРМ, хосты, оборудование, МФУ, телефоны)
var referencesJS = `
// ==================== ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ ====================
// Функция очистки значений для корректного отображения
function cleanDisplayValue(value) {
    if (value === null || value === undefined || value === '' || 
        value === 'null' || value === 'NULL' || value === '<nil>' || value === 'nil') {
        return '-';
    }
    return value;
}

// ==================== ЗАГРУЗКА И ОТОБРАЖЕНИЕ СПРАВОЧНИКОВ ====================
function renderAddressesTable() {
	if (!addressesTable) return;
	if (addresses.length === 0) {
		addressesTable.innerHTML = '<tr><td colspan="7" class="empty-state">Адреса не найдены</td></tr>';
		return;
	}
	addresses.sort((a, b) => {
		if (a.street !== b.street) return a.street.localeCompare(b.street);
		if (a.building !== b.building) return a.building.localeCompare(b.building);
		if (a.floor !== b.floor) return (a.floor || 0) - (b.floor || 0);
		if (a.cabinet !== b.cabinet) return (a.cabinet || '').localeCompare(b.cabinet || '');
		if (a.corridor !== b.corridor) return (a.corridor || '').localeCompare(b.corridor || '');
		return (a.service_room || '').localeCompare(b.service_room || '');
	});
	var html = '';
	addresses.forEach(addr => {
		html += '<tr data-id="' + addr.id + '">' +
			'<td>' + escapeHtml(addr.street) + '</td>' +
			'<td>' + escapeHtml(addr.building) + '</td>' +
			'<td>' + (addr.cabinet || '-') + '</td>' +
			'<td>' + (addr.corridor || '-') + '</td>' +
			'<td>' + (addr.floor || '-') + '</td>' +
			'<td>' + (addr.service_room || '-') + '</td>' +
			'<td>' +
				'<button class="btn btn-sm btn-edit edit-btn" onclick="editAddressRow(' + addr.id + ', this)">✏️</button>' +
				'<button class="btn btn-sm btn-save edit-btn" onclick="saveAddressRow(' + addr.id + ', this)" style="display:none;">💾</button>' +
				'<button class="btn btn-sm btn-cancel edit-btn" onclick="cancelAddressEdit(' + addr.id + ', this)" style="display:none;">❌</button>' +
				'<button class="btn btn-sm btn-danger" onclick="deleteAddress(' + addr.id + ')">Удалить</button>' +
			'</td>' +
		'</tr>';
	});
	addressesTable.innerHTML = html;
}

function renderEmployeesTable() {
	if (!employeesTable) return;
	if (employees.length === 0) {
		employeesTable.innerHTML = '<tr><td colspan="6" class="empty-state">Сотрудники не найдены</td></tr>';
		return;
	}
	var html = '';
	employees.forEach(emp => {
		html += '<tr data-id="' + emp.id + '">' +
			'<td>' + escapeHtml(emp.full_name) + '</td>' +
			'<td>' + (emp.short_name || '-') + '</td>' +
			'<td>' + (emp.phone_city || '-') + '</td>' +
			'<td>' + (emp.phone_internal || '-') + '</td>' +
			'<td>' + escapeHtml(emp.full_address) + '</td>' +
			'<td>' +
				'<button class="btn btn-sm btn-edit edit-btn" onclick="editEmployeeRow(' + emp.id + ', this)">✏️</button>' +
				'<button class="btn btn-sm btn-save edit-btn" onclick="saveEmployeeRow(' + emp.id + ', this)" style="display:none;">💾</button>' +
				'<button class="btn btn-sm btn-cancel edit-btn" onclick="cancelEmployeeEdit(' + emp.id + ', this)" style="display:none;">❌</button>' +
				'<button class="btn btn-sm btn-danger" onclick="deleteEmployee(' + emp.id + ')">Удалить</button>' +
			'</td>' +
		'</tr>';
	});
	employeesTable.innerHTML = html;
}

// Глобальные переменные для фильтрации и сортировки АРМ
var filteredWorkstations = [];
var workstationSortColumn = '';
var workstationSortDirection = 'asc';

// Глобальные переменные для фильтрации и сортировки адресов
var filteredAddresses = [];
var addressSortColumn = '';
var addressSortDirection = 'asc';

// Глобальные переменные для фильтрации и сортировки сотрудников
var filteredEmployees = [];
var employeeSortColumn = '';
var employeeSortDirection = 'asc';

// Глобальные переменные для фильтрации и сортировки хостов
var filteredHosts = [];
var hostSortColumn = '';
var hostSortDirection = 'asc';

// Глобальные переменные для фильтрации и сортировки сетевого оборудования
var filteredNetworkEquipment = [];
var networkEquipmentSortColumn = '';
var networkEquipmentSortDirection = 'asc';

// Глобальные переменные для фильтрации и сортировки МФУ
var filteredNetworkMFPs = [];
var networkMFPSortColumn = '';
var networkMFPSortDirection = 'asc';

// Глобальные переменные для фильтрации и сортировки IP-телефонов
var filteredIPPhones = [];
var ipPhoneSortColumn = '';
var ipPhoneSortDirection = 'asc';

// ==================== ФУНКЦИИ ДЛЯ СПРАВОЧНИКА АДРЕСОВ ====================
function filterAddressesTable() {
	filteredAddresses = addresses.filter(function(addr) {
		var inputs = document.querySelectorAll('#addresses-table .search-row input');
		for (var i = 0; i < inputs.length; i++) {
			var query = inputs[i].value.toLowerCase().trim();
			if (query) {
				var value = '';
				switch(i) {
					case 0: value = addr.street || ''; break;
					case 1: value = addr.building || ''; break;
					case 2: value = addr.cabinet || ''; break;
					case 3: value = addr.corridor || ''; break;
					case 4: value = (addr.floor || '').toString(); break;
					case 5: value = addr.service_room || ''; break;
				}
				if (value.toLowerCase().indexOf(query) === -1) {
					return false;
				}
			}
		}
		return true;
	});
	
	var pageSize = parseInt(document.getElementById('addresses-page-size').value) || 25;
	var totalPages = pageSize === 0 ? 1 : Math.ceil(filteredAddresses.length / pageSize);
	updateAddressesPagination(1, totalPages);
	renderAddressesTable();
}

function sortAddressesTable(column) {
	if (addressSortColumn === column) {
		addressSortDirection = addressSortDirection === 'asc' ? 'desc' : 'asc';
	} else {
		addressSortColumn = column;
		addressSortDirection = 'asc';
	}
	renderAddressesTable();
}

function updateAddressesPagination(currentPage, totalPages) {
	var pagination = document.getElementById('addresses-pagination');
	if (!pagination) return;
	
	pagination.innerHTML = 
		'<li class="page-item ' + (currentPage === 1 ? 'disabled' : '') + '">' +
			'<a class="page-link" href="#" onclick="changeAddressesPage(event, \'prev\')" aria-label="Previous">' +
				'<span aria-hidden="true">&laquo;</span>' +
			'</a>' +
		'</li>';
	
	for (var i = 1; i <= totalPages && i <= 5; i++) {
		pagination.innerHTML += 
			'<li class="page-item ' + (i === currentPage ? 'active' : '') + '">' +
				'<a class="page-link" href="#" onclick="changeAddressesPage(event, ' + i + ')">' + i + '</a>' +
			'</li>';
	}
	
	pagination.innerHTML += 
		'<li class="page-item ' + (currentPage === totalPages ? 'disabled' : '') + '">' +
			'<a class="page-link" href="#" onclick="changeAddressesPage(event, \'next\')" aria-label="Next">' +
				'<span aria-hidden="true">&raquo;</span>' +
			'</a>' +
		'</li>';
}

function changeAddressesPage(event, direction) {
	if (event && event.preventDefault) {
		event.preventDefault();
	}
	
	var currentPage = parseInt(document.querySelector('#addresses-pagination .page-item.active .page-link').textContent) || 1;
	var pageSize = parseInt(document.getElementById('addresses-page-size').value) || 25;
	var totalItems = filteredAddresses.length > 0 ? filteredAddresses.length : addresses.length;
	var totalPages = pageSize === 0 ? 1 : Math.ceil(totalItems / pageSize);
	
	var newPage;
	if (direction === 'prev') {
		newPage = Math.max(1, currentPage - 1);
	} else if (direction === 'next') {
		newPage = Math.min(totalPages, currentPage + 1);
	} else if (typeof direction === 'number') {
		newPage = Math.max(1, Math.min(totalPages, direction));
	}
	
	if (newPage) {
		updateAddressesPagination(newPage, totalPages);
		renderAddressesTable();
	}
}

function renderAddressesTable() {
	if (!addressesTable) return;
	
	var container = document.querySelector('.scrollable-table-container');
	var scrollTop = container ? container.scrollTop : 0;
	
	var dataToRender = filteredAddresses.length > 0 ? filteredAddresses : addresses;
	
	if (dataToRender.length === 0) {
		addressesTable.innerHTML = '<tr><td colspan=\"7\" class=\"empty-state\">Адреса не найдены</td></tr>';
		var infoEl = document.getElementById('addresses-info');
		if (infoEl) infoEl.textContent = 'Показано 0 из ' + addresses.length + ' записей';
		
		if (container) {
			container.scrollTop = scrollTop;
		}
		return;
	}
	
	var sorted = [...dataToRender];
	if (addressSortColumn) {
		sorted.sort(function(a, b) {
			var valA = a[addressSortColumn] || '';
			var valB = b[addressSortColumn] || '';
			
			if (typeof valA === 'string' && typeof valB === 'string') {
				valA = valA.toLowerCase();
				valB = valB.toLowerCase();
			}
			
			if (valA < valB) return addressSortDirection === 'asc' ? -1 : 1;
			if (valA > valB) return addressSortDirection === 'asc' ? 1 : -1;
			return 0;
		});
	}
	
	var pageSize = parseInt(document.getElementById('addresses-page-size').value) || 25;
	var currentPage = parseInt(document.querySelector('#addresses-pagination .page-item.active .page-link').textContent) || 1;
	var totalPages = pageSize === 0 ? 1 : Math.ceil(sorted.length / pageSize);
	
	updateAddressesPagination(currentPage, totalPages);
	
	if (currentPage > totalPages && totalPages > 0) {
		currentPage = totalPages;
		updateAddressesPagination(currentPage, totalPages);
	}
	
	var start = pageSize === 0 ? 0 : (currentPage - 1) * pageSize;
	var end = pageSize === 0 ? sorted.length : start + pageSize;
	var pageItems = pageSize === 0 ? sorted : sorted.slice(start, end);
	
	var html = '';
	pageItems.forEach(addr => {
		html += '<tr data-id="' + addr.id + '">' +
			'<td>' + escapeHtml(addr.street) + '</td>' +
			'<td>' + escapeHtml(addr.building) + '</td>' +
			'<td>' + (addr.cabinet || '-') + '</td>' +
			'<td>' + (addr.corridor || '-') + '</td>' +
			'<td>' + (addr.floor || '-') + '</td>' +
			'<td>' + (addr.service_room || '-') + '</td>' +
			'<td>' +
				'<button class="btn btn-sm btn-edit edit-btn" onclick="editAddressRow(' + addr.id + ', this)">✏️</button>' +
				'<button class="btn btn-sm btn-save edit-btn" onclick="saveAddressRow(' + addr.id + ', this)" style="display:none;">💾</button>' +
				'<button class="btn btn-sm btn-cancel edit-btn" onclick="cancelAddressEdit(' + addr.id + ', this)" style="display:none;">❌</button>' +
				'<button class="btn btn-sm btn-danger" onclick="deleteAddress(' + addr.id + ')">Удалить</button>' +
			'</td>' +
		'</tr>';
	});
	
	addressesTable.innerHTML = html;
	
	var infoEl = document.getElementById('addresses-info');
	if (infoEl) {
		infoEl.textContent = 'Показано ' + (pageItems.length === 0 ? 0 : start + 1) + '-' + end + ' из ' + sorted.length + ' записей (всего: ' + addresses.length + ')';
	}
	
	if (container) {
		container.scrollTop = scrollTop;
	}
}

// ==================== ФУНКЦИИ ДЛЯ СПРАВОЧНИКА СОТРУДНИКОВ ====================
function filterEmployeesTable() {
	filteredEmployees = employees.filter(function(emp) {
		var inputs = document.querySelectorAll('#employees-table .search-row input');
		for (var i = 0; i < inputs.length; i++) {
			var query = inputs[i].value.toLowerCase().trim();
			if (query) {
				var value = '';
				switch(i) {
					case 0: value = emp.full_name || ''; break;
					case 1: value = emp.short_name || ''; break;
					case 2: value = emp.phone_city || ''; break;
					case 3: value = emp.phone_internal || ''; break;
					case 4: value = emp.full_address || ''; break;
				}
				if (value.toLowerCase().indexOf(query) === -1) {
					return false;
				}
			}
		}
		return true;
	});
	
	var pageSize = parseInt(document.getElementById('employees-page-size').value) || 25;
	var totalPages = pageSize === 0 ? 1 : Math.ceil(filteredEmployees.length / pageSize);
	updateEmployeesPagination(1, totalPages);
	renderEmployeesTable();
}

function sortEmployeesTable(column) {
	if (employeeSortColumn === column) {
		employeeSortDirection = employeeSortDirection === 'asc' ? 'desc' : 'asc';
	} else {
		employeeSortColumn = column;
		employeeSortDirection = 'asc';
	}
	renderEmployeesTable();
}

function updateEmployeesPagination(currentPage, totalPages) {
	var pagination = document.getElementById('employees-pagination');
	if (!pagination) return;
	
	pagination.innerHTML = 
		'<li class="page-item ' + (currentPage === 1 ? 'disabled' : '') + '">' +
			'<a class="page-link" href="#" onclick="changeEmployeesPage(event, \'prev\')" aria-label="Previous">' +
				'<span aria-hidden="true">&laquo;</span>' +
			'</a>' +
		'</li>';
	
	for (var i = 1; i <= totalPages && i <= 5; i++) {
		pagination.innerHTML += 
			'<li class="page-item ' + (i === currentPage ? 'active' : '') + '">' +
				'<a class="page-link" href="#" onclick="changeEmployeesPage(event, ' + i + ')">' + i + '</a>' +
			'</li>';
	}
	
	pagination.innerHTML += 
		'<li class="page-item ' + (currentPage === totalPages ? 'disabled' : '') + '">' +
			'<a class="page-link" href="#" onclick="changeEmployeesPage(event, \'next\')" aria-label="Next">' +
				'<span aria-hidden="true">&raquo;</span>' +
			'</a>' +
		'</li>';
}

function changeEmployeesPage(event, direction) {
	if (event && event.preventDefault) {
		event.preventDefault();
	}
	
	var currentPage = parseInt(document.querySelector('#employees-pagination .page-item.active .page-link').textContent) || 1;
	var pageSize = parseInt(document.getElementById('employees-page-size').value) || 25;
	var totalItems = filteredEmployees.length > 0 ? filteredEmployees.length : employees.length;
	var totalPages = pageSize === 0 ? 1 : Math.ceil(totalItems / pageSize);
	
	var newPage;
	if (direction === 'prev') {
		newPage = Math.max(1, currentPage - 1);
	} else if (direction === 'next') {
		newPage = Math.min(totalPages, currentPage + 1);
	} else if (typeof direction === 'number') {
		newPage = Math.max(1, Math.min(totalPages, direction));
	}
	
	if (newPage) {
		updateEmployeesPagination(newPage, totalPages);
		renderEmployeesTable();
	}
}

function renderEmployeesTable() {
	if (!employeesTable) return;
	
	var container = document.querySelector('.scrollable-table-container');
	var scrollTop = container ? container.scrollTop : 0;
	
	var dataToRender = filteredEmployees.length > 0 ? filteredEmployees : employees;
	
	if (dataToRender.length === 0) {
		employeesTable.innerHTML = '<tr><td colspan=\"6\" class=\"empty-state\">Сотрудники не найдены</td></tr>';
		var infoEl = document.getElementById('employees-info');
		if (infoEl) infoEl.textContent = 'Показано 0 из ' + employees.length + ' записей';
		
		if (container) {
			container.scrollTop = scrollTop;
		}
		return;
	}
	
	var sorted = [...dataToRender];
	if (employeeSortColumn) {
		sorted.sort(function(a, b) {
			var valA = a[employeeSortColumn] || '';
			var valB = b[employeeSortColumn] || '';
			
			if (typeof valA === 'string' && typeof valB === 'string') {
				valA = valA.toLowerCase();
				valB = valB.toLowerCase();
			}
			
			if (valA < valB) return employeeSortDirection === 'asc' ? -1 : 1;
			if (valA > valB) return employeeSortDirection === 'asc' ? 1 : -1;
			return 0;
		});
	}
	
	var pageSize = parseInt(document.getElementById('employees-page-size').value) || 25;
	var currentPage = parseInt(document.querySelector('#employees-pagination .page-item.active .page-link').textContent) || 1;
	var totalPages = pageSize === 0 ? 1 : Math.ceil(sorted.length / pageSize);
	
	updateEmployeesPagination(currentPage, totalPages);
	
	if (currentPage > totalPages && totalPages > 0) {
		currentPage = totalPages;
		updateEmployeesPagination(currentPage, totalPages);
	}
	
	var start = pageSize === 0 ? 0 : (currentPage - 1) * pageSize;
	var end = pageSize === 0 ? sorted.length : start + pageSize;
	var pageItems = pageSize === 0 ? sorted : sorted.slice(start, end);
	
	var html = '';
	pageItems.forEach(emp => {
		html += '<tr data-id="' + emp.id + '">' +
			'<td>' + escapeHtml(emp.full_name) + '</td>' +
			'<td>' + (emp.short_name || '-') + '</td>' +
			'<td>' + (emp.phone_city || '-') + '</td>' +
			'<td>' + (emp.phone_internal || '-') + '</td>' +
			'<td>' + escapeHtml(emp.full_address) + '</td>' +
			'<td>' +
				'<button class="btn btn-sm btn-edit edit-btn" onclick="editEmployeeRow(' + emp.id + ', this)">✏️</button>' +
				'<button class="btn btn-sm btn-save edit-btn" onclick="saveEmployeeRow(' + emp.id + ', this)" style="display:none;">💾</button>' +
				'<button class="btn btn-sm btn-cancel edit-btn" onclick="cancelEmployeeEdit(' + emp.id + ', this)" style="display:none;">❌</button>' +
				'<button class="btn btn-sm btn-danger" onclick="deleteEmployee(' + emp.id + ')">Удалить</button>' +
			'</td>' +
		'</tr>';
	});
	
	employeesTable.innerHTML = html;
	
	var infoEl = document.getElementById('employees-info');
	if (infoEl) {
		infoEl.textContent = 'Показано ' + (pageItems.length === 0 ? 0 : start + 1) + '-' + end + ' из ' + sorted.length + ' записей (всего: ' + employees.length + ')';
	}
	
	if (container) {
		container.scrollTop = scrollTop;
	}
}

// Функция фильтрации таблицы АРМ
function filterWorkstationsTable() {
	filteredWorkstations = workstations.filter(function(ws) {
		var inputs = document.querySelectorAll('#workstations-table .search-row input');
		for (var i = 0; i < inputs.length - 1; i++) {
			var query = inputs[i].value.toLowerCase().trim();
			if (query) {
				var value = '';
				switch(i) {
					case 0: value = ws.full_address || ''; break;
					case 1: value = ws.employee_short_name || ''; break;
					case 2: value = ws.inventory_number || ''; break;
					case 3: value = ws.serial_number || ''; break;
					case 4: 
						try {
							var seals = JSON.parse(ws.seal_numbers || '[]');
							value = seals.join(', ');
						} catch(e) { value = ws.seal_numbers || ''; }
						break;
					case 5: value = ws.monitor_count.toString(); break;
					case 6: 
						value = (ws.is_vacant === '1' || ws.is_vacant === 1 || ws.is_vacant === true) ? 'вакантное' : 'действующее';
						break;
					case 7: value = ws.replacement_date || ''; break;
					case 8: value = ws.replacement_letter || ''; break;
				}
				if (value.toLowerCase().indexOf(query) === -1) {
					return false;
				}
			}
		}
		return true;
	});
	
	// Сбрасываем на первую страницу при фильтрации
	var pageSize = parseInt(document.getElementById('workstations-page-size').value) || 25;
	var totalPages = pageSize === 0 ? 1 : Math.ceil(filteredWorkstations.length / pageSize);
	updateWorkstationsPagination(1, totalPages);
	renderWorkstationsTable();
}

// Функция сортировки таблицы АРМ
function sortWorkstationsTable(column) {
	if (workstationSortColumn === column) {
		workstationSortDirection = workstationSortDirection === 'asc' ? 'desc' : 'asc';
	} else {
		workstationSortColumn = column;
		workstationSortDirection = 'asc';
	}
	renderWorkstationsTable();
}

// Функция обновления пагинации (ИСПРАВЛЕНА: всегда обновлять элементы)
function updateWorkstationsPagination(currentPage, totalPages) {
	var pagination = document.getElementById('workstations-pagination');
	if (!pagination) return;
	
	// ИСПРАВЛЕНО: всегда перерисовываем пагинацию
	pagination.innerHTML = 
		'<li class="page-item ' + (currentPage === 1 ? 'disabled' : '') + '">' +
			'<a class="page-link" href="#" onclick="changeWorkstationsPage(event, \'prev\')" aria-label="Previous">' +
				'<span aria-hidden="true">&laquo;</span>' +
			'</a>' +
		'</li>';
	
	for (var i = 1; i <= totalPages && i <= 5; i++) {
		pagination.innerHTML += 
			'<li class="page-item ' + (i === currentPage ? 'active' : '') + '">' +
				'<a class="page-link" href="#" onclick="changeWorkstationsPage(event, ' + i + ')">' + i + '</a>' +
			'</li>';
	}
	
	pagination.innerHTML += 
		'<li class="page-item ' + (currentPage === totalPages ? 'disabled' : '') + '">' +
			'<a class="page-link" href="#" onclick="changeWorkstationsPage(event, \'next\')" aria-label="Next">' +
				'<span aria-hidden="true">&raquo;</span>' +
			'</a>' +
		'</li>';
}

// Функция смены страницы (ИСПРАВЛЕНА: всегда обновлять таблицу)
function changeWorkstationsPage(event, direction) {
	// ИСПРАВЛЕНО: безопасная проверка события
	if (event && event.preventDefault) {
		event.preventDefault();
	}
	
	var currentPage = parseInt(document.querySelector('#workstations-pagination .page-item.active .page-link').textContent) || 1;
	var pageSize = parseInt(document.getElementById('workstations-page-size').value) || 25;
	var totalItems = filteredWorkstations.length > 0 ? filteredWorkstations.length : workstations.length;
	var totalPages = pageSize === 0 ? 1 : Math.ceil(totalItems / pageSize);
	
	var newPage;
	if (direction === 'prev') {
		newPage = Math.max(1, currentPage - 1);
	} else if (direction === 'next') {
		newPage = Math.min(totalPages, currentPage + 1);
	} else if (typeof direction === 'number') {
		newPage = Math.max(1, Math.min(totalPages, direction));
	}
	
	// ИСПРАВЛЕНО: всегда обновляем пагинацию и таблицу
	// (необходимо при смене размера страницы)
	if (newPage) {
		updateWorkstationsPagination(newPage, totalPages);
		renderWorkstationsTable();
	}
}

function renderWorkstationsTable() {
	if (!workstationsTable) return;
	
	// ИСПРАВЛЕНО: сохраняем позицию прокрутки контейнера перед обновлением
	var container = document.querySelector('.scrollable-table-container');
	var scrollTop = container ? container.scrollTop : 0;
	
	// Определяем данные для рендеринга (фильтрованные или все)
	var dataToRender = filteredWorkstations.length > 0 ? filteredWorkstations : workstations;
	
	if (dataToRender.length === 0) {
		document.getElementById('workstations-table-body').innerHTML = '<tr><td colspan="10" class="empty-state">АРМ не найдены</td></tr>';
		document.getElementById('workstations-info').textContent = 'Показано 0 из ' + workstations.length + ' записей';
		
		// ИСПРАВЛЕНО: восстанавливаем прокрутку
		if (container) {
			container.scrollTop = scrollTop;
		}
		return;
	}
	
	// Сортировка
	var sorted = [...dataToRender];
	if (workstationSortColumn) {
		sorted.sort(function(a, b) {
			var valA, valB;
			
			if (workstationSortColumn === 'is_vacant') {
				valA = a.is_vacant === '1' || a.is_vacant === 1 || a.is_vacant === true;
				valB = b.is_vacant === '1' || b.is_vacant === 1 || b.is_vacant === true;
			} else {
				valA = a[workstationSortColumn] || '';
				valB = b[workstationSortColumn] || '';
			}
			
			if (typeof valA === 'string' && typeof valB === 'string') {
				valA = valA.toLowerCase();
				valB = valB.toLowerCase();
			}
			
			if (valA < valB) return workstationSortDirection === 'asc' ? -1 : 1;
			if (valA > valB) return workstationSortDirection === 'asc' ? 1 : -1;
			return 0;
		});
	}
	
	// Пагинация
	var pageSize = parseInt(document.getElementById('workstations-page-size').value) || 25;
	var currentPage = parseInt(document.querySelector('#workstations-pagination .page-item.active .page-link').textContent) || 1;
	var totalPages = pageSize === 0 ? 1 : Math.ceil(sorted.length / pageSize);
	
	// ИСПРАВЛЕНО: всегда обновляем пагинацию при рендеринге
	updateWorkstationsPagination(currentPage, totalPages);
	
	if (currentPage > totalPages && totalPages > 0) {
		currentPage = totalPages;
		updateWorkstationsPagination(currentPage, totalPages);
	}
	
	var start = pageSize === 0 ? 0 : (currentPage - 1) * pageSize;
	var end = pageSize === 0 ? sorted.length : start + pageSize;
	var pageItems = pageSize === 0 ? sorted : sorted.slice(start, end);
	
	// Рендеринг
	var html = '';
	pageItems.forEach(function(ws) {
		var status = ws.is_vacant === '1' || ws.is_vacant === 1 || ws.is_vacant === true ? 
			'<span style="color:#FFA500;font-weight:bold;">Вакантное</span>' : 
			'<span style="color:#3F8C44;font-weight:bold;">Действующее</span>';
		
		// Номера пломб
		var sealNumbers = '-';
		try {
			var seals = JSON.parse(ws.seal_numbers || '[]');
			sealNumbers = seals.length > 0 ? seals.join(', ') : '-';
		} catch(e) {
			sealNumbers = cleanDisplayValue(ws.seal_numbers);
		}
		
		// ИСПРАВЛЕНО: корректная обработка полей замены оборудования
		var replacementDate = cleanDisplayValue(ws.replacement_date);
		var replacementLetter = cleanDisplayValue(ws.replacement_letter);
		
		html += '<tr data-id="' + ws.id + '">' +
			'<td>' + escapeHtml(ws.full_address) + '</td>' +
			'<td>' + (ws.employee_short_name || '-') + '</td>' +
			'<td>' + escapeHtml(ws.inventory_number) + '</td>' +
			'<td>' + escapeHtml(ws.serial_number || '-') + '</td>' +
			'<td>' + escapeHtml(sealNumbers) + '</td>' +
			'<td>' + ws.monitor_count + '</td>' +
			'<td>' + status + '</td>' +
			'<td>' + replacementDate + '</td>' +
			'<td>' + replacementLetter + '</td>' +
			'<td>' +
				'<button class="btn btn-sm btn-edit edit-btn" onclick="editWorkstationRow(' + ws.id + ', this)">✏️</button> ' +
				'<button class="btn btn-sm btn-danger" onclick="deleteWorkstation(' + ws.id + ')">🗑️</button>' +
			'</td>' +
		'</tr>';
	});
	
	document.getElementById('workstations-table-body').innerHTML = html;
	document.getElementById('workstations-info').textContent = 
		'Показано ' + (pageItems.length === 0 ? 0 : start + 1) + '-' + end + ' из ' + sorted.length + ' записей (всего: ' + workstations.length + ')';
	
	// ИСПРАВЛЕНО: восстанавливаем прокрутку контейнера после обновления
	if (container) {
		container.scrollTop = scrollTop;
	}
}

function renderHostsTable() {
	if (!hostsTable) return;
	if (hosts.length === 0) {
		hostsTable.innerHTML = '<tr><td colspan="6" class="empty-state">Хосты не найдены</td></tr>';
		return;
	}
	hosts.sort((a, b) => {
		if (a.full_address !== b.full_address) return a.full_address.localeCompare(b.full_address);
		return (a.employee_short_name || '').localeCompare(b.employee_short_name || '');
	});
	var html = '';
	hosts.forEach(host => {
		var statusClass = host.enabled === '1' || host.enabled === 1 || host.enabled === true ? 'success' : 'danger';
		var statusText = host.enabled === '1' || host.enabled === 1 || host.enabled === true ? 'Включён' : 'Отключён';
		html += '<tr data-id="' + host.id + '">' +
			'<td>' + escapeHtml(host.full_address) + '</td>' +
			'<td>' + (host.employee_short_name || '-') + '</td>' +
			'<td><code>' + escapeHtml(host.ip) + '</code></td>' +
			'<td>' + host.ssh_port + '</td>' +
			'<td><span class="' + statusClass + '">' + statusText + '</span></td>' +
			'<td>' +
				'<button class="btn btn-sm btn-edit edit-btn" onclick="editHostRow(' + host.id + ', this)">✏️</button>' +
				'<button class="btn btn-sm btn-save edit-btn" onclick="saveHostRow(' + host.id + ', this)" style="display:none;">💾</button>' +
				'<button class="btn btn-sm btn-cancel edit-btn" onclick="cancelHostEdit(' + host.id + ', this)" style="display:none;">❌</button>' +
				'<button class="btn btn-sm btn-danger" onclick="deleteHost(' + host.id + ')">Удалить</button>' +
			'</td>' +
		'</tr>';
	});
	hostsTable.innerHTML = html;
}

function renderNetworkEquipmentTable() {
	if (!networkEquipmentTable) return;
	if (networkEquipment.length === 0) {
		networkEquipmentTable.innerHTML = '<tr><td colspan="6" class="empty-state">Оборудование не найдено</td></tr>';
		return;
	}
	networkEquipment.sort((a, b) => {
		if (a.full_address !== b.full_address) return a.full_address.localeCompare(b.full_address);
		if (a.category !== b.category) return a.category.localeCompare(b.category);
		return a.model.localeCompare(b.model);
	});
	var html = '';
	networkEquipment.forEach(eq => {
		html += '<tr data-id="' + eq.id + '">' +
			'<td>' + escapeHtml(eq.full_address) + '</td>' +
			'<td>' + escapeHtml(eq.category) + '</td>' +
			'<td>' + escapeHtml(eq.type) + '</td>' +
			'<td>' + escapeHtml(eq.model) + '</td>' +
			'<td>' + eq.port_count + '</td>' +
			'<td>' +
				'<button class="btn btn-sm btn-edit edit-btn" onclick="editNetworkEquipmentRow(' + eq.id + ', this)">✏️</button>' +
				'<button class="btn btn-sm btn-save edit-btn" onclick="saveNetworkEquipmentRow(' + eq.id + ', this)" style="display:none;">💾</button>' +
				'<button class="btn btn-sm btn-cancel edit-btn" onclick="cancelNetworkEquipmentEdit(' + eq.id + ', this)" style="display:none;">❌</button>' +
				'<button class="btn btn-sm btn-danger" onclick="deleteNetworkEquipment(' + eq.id + ')">Удалить</button>' +
			'</td>' +
		'</tr>';
	});
	networkEquipmentTable.innerHTML = html;
}

function renderNetworkMFPsTable() {
	if (!networkMFPsTable) return;
	if (networkMFPs.length === 0) {
		networkMFPsTable.innerHTML = '<tr><td colspan="6" class="empty-state">МФУ не найдены</td></tr>';
		return;
	}
	networkMFPs.sort((a, b) => {
		if (a.full_address !== b.full_address) return a.full_address.localeCompare(b.full_address);
		return a.model.localeCompare(b.model);
	});
	var html = '';
	networkMFPs.forEach(mfp => {
		html += '<tr data-id="' + mfp.id + '">' +
			'<td>' + escapeHtml(mfp.full_address) + '</td>' +
			'<td>' + escapeHtml(mfp.model) + '</td>' +
			'<td><code>' + escapeHtml(mfp.ip) + '</code></td>' +
			'<td>' + (mfp.hostname || '-') + '</td>' +
			'<td>' + escapeHtml(mfp.serial_number) + '</td>' +
			'<td>' +
				'<button class="btn btn-sm btn-edit edit-btn" onclick="editNetworkMFPRow(' + mfp.id + ', this)">✏️</button>' +
				'<button class="btn btn-sm btn-save edit-btn" onclick="saveNetworkMFPRow(' + mfp.id + ', this)" style="display:none;">💾</button>' +
				'<button class="btn btn-sm btn-cancel edit-btn" onclick="cancelNetworkMFPEdit(' + mfp.id + ', this)" style="display:none;">❌</button>' +
				'<button class="btn btn-sm btn-danger" onclick="deleteNetworkMFP(' + mfp.id + ')">Удалить</button>' +
			'</td>' +
		'</tr>';
	});
	networkMFPsTable.innerHTML = html;
}

function renderIPPhonesTable() {
	if (!ipPhonesTable) return;
	if (ipPhones.length === 0) {
		ipPhonesTable.innerHTML = '<tr><td colspan="4" class="empty-state">Телефоны не найдены</td></tr>';
		return;
	}
	ipPhones.sort((a, b) => {
		if (a.full_address !== b.full_address) return a.full_address.localeCompare(b.full_address);
		return a.employee_full_name.localeCompare(b.employee_full_name);
	});
	var html = '';
	ipPhones.forEach(phone => {
		html += '<tr data-id="' + phone.id + '">' +
			'<td>' + escapeHtml(phone.full_address) + '</td>' +
			'<td>' + escapeHtml(phone.employee_full_name) + '</td>' +
			'<td><code>' + escapeHtml(phone.ip) + '</code></td>' +
			'<td>' +
				'<button class="btn btn-sm btn-edit edit-btn" onclick="editIPPhoneRow(' + phone.id + ', this)">✏️</button>' +
				'<button class="btn btn-sm btn-save edit-btn" onclick="saveIPPhoneRow(' + phone.id + ', this)" style="display:none;">💾</button>' +
				'<button class="btn btn-sm btn-cancel edit-btn" onclick="cancelIPPhoneEdit(' + phone.id + ', this)" style="display:none;">❌</button>' +
				'<button class="btn btn-sm btn-danger" onclick="deleteIPPhone(' + phone.id + ')">Удалить</button>' +
			'</td>' +
		'</tr>';
	});
	ipPhonesTable.innerHTML = html;
}

// ==================== ФУНКЦИИ ЗАПОЛНЕНИЯ ВЫПАДАЮЩИХ СПИСКОВ ====================
function populateAddressDropdowns() {
	if (!addressStreet) return;
	
	var addressSelects = document.querySelectorAll('select[id$="-address"], select[name="address"]');
	addressSelects.forEach(select => {
		select.innerHTML = '<option value="">-- Выберите адрес --</option>';
		addresses.forEach(addr => {
			var option = document.createElement('option');
			option.value = addr.id;
			option.textContent = addr.full_address;
			select.appendChild(option);
		});
	});
}

function populateEmployeeDropdowns() {
	if (!employeeFullName) return;
	
	var employeeSelects = document.querySelectorAll('select[id$="-employee"], select[name="employee"]');
	employeeSelects.forEach(select => {
		select.innerHTML = '<option value="">-- Выберите сотрудника --</option>';
		employees.forEach(emp => {
			var option = document.createElement('option');
			option.value = emp.id;
			option.textContent = emp.short_name || emp.full_name;
			select.appendChild(option);
		});
	});
}

function populateWorkstationDropdowns() {
	if (!workstationAddress) return;
	
	workstationAddress.innerHTML = '<option value="">-- Выберите адрес --</option>';
	addresses.forEach(addr => {
		var option = document.createElement('option');
		option.value = addr.id;
		option.textContent = addr.full_address;
		workstationAddress.appendChild(option);
	});
	
	workstationEmployee.innerHTML = '<option value="">-- Выберите сотрудника --</option>';
	employees.forEach(emp => {
		var option = document.createElement('option');
		option.value = emp.id;
		option.textContent = emp.short_name || emp.full_name;
		workstationEmployee.appendChild(option);
	});
}

// ==================== ФУНКЦИИ СПРАВОЧНИКОВ (CRUD) ====================
function saveAddress() {
	var street = addressStreet.value.trim();
	var building = addressBuilding.value.trim();
	var cabinet = addressCabinet.value.trim();
	var corridor = addressCorridor.value.trim();
	var floor = addressFloor.value;
	var serviceRoom = addressServiceRoom.value;

	if (!street) {
		showAlert(addressAlert, 'Улица обязательна для заполнения', 'danger');
		return;
	}
	if (!building) {
		showAlert(addressAlert, 'Дом обязателен для заполнения', 'danger');
		return;
	}

	var addressType = document.querySelector('input[name="address-type"]:checked').value;
	if (addressType === 'cabinet' && !cabinet) {
		showAlert(addressAlert, 'Номер кабинета обязателен для заполнения', 'danger');
		return;
	}
	if (addressType === 'corridor') {
		if (!corridor) {
			showAlert(addressAlert, 'Название коридора обязательно для заполнения', 'danger');
			return;
		}
		if (!floor) {
			showAlert(addressAlert, 'Этаж обязателен для заполнения при выборе коридора', 'danger');
			return;
		}
	}
	if (addressType === 'service' && !serviceRoom) {
		showAlert(addressAlert, 'Служебное помещение обязательно для выбора', 'danger');
		return;
	}

	var data = {
		street: street,
		building: building
	};
	if (addressType === 'cabinet') {
		data.cabinet = cabinet;
	} else if (addressType === 'corridor') {
		data.corridor = corridor;
		data.floor = parseInt(floor);
	} else if (addressType === 'service') {
		data.service_room = serviceRoom;
	}

	fetch('/api/reference/addresses', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(data)
	})
	.then(response => response.json())
	.then(result => {
		addressStreet.value = '';
		addressBuilding.value = '';
		addressCabinet.value = '';
		addressCorridor.value = '';
		addressFloor.value = '';
		addressServiceRoom.value = '';
		addressTypeCabinet.checked = true;
		toggleAddressType();

		fetch('/api/reference/addresses')
			.then(r => r.json())
			.then(data => {
				addresses = data;
				renderAddressesTable();
				populateAddressDropdowns();
				populateWorkstationDropdowns();
				showAlert(null, 'Адрес успешно добавлен!', 'success', true);
			});
	})
	.catch(error => {
		showAlert(addressAlert, error.message || 'Ошибка сети', 'danger');
	});
}

function deleteAddress(id) {
	if (!confirm('Вы уверены, что хотите удалить этот адрес?')) {
		return;
	}
	fetch('/api/reference/addresses/' + id, {
		method: 'DELETE'
	})
	.then(response => response.json())
	.then(result => {
		fetch('/api/reference/addresses')
			.then(r => r.json())
			.then(data => {
				addresses = data;
				renderAddressesTable();
				populateAddressDropdowns();
				populateWorkstationDropdowns();
				showAlert(null, 'Адрес успешно удалён!', 'success', true);
			});
	})
	.catch(error => {
		showAlert(null, error.message || 'Ошибка удаления адреса', 'danger', true);
	});
}

function editAddressRow(id, button) {
	var row = button.closest('tr');
	var cells = row.querySelectorAll('td');

	var originalValues = [];
	for (var i = 0; i < cells.length - 1; i++) {
		originalValues.push(cells[i].innerHTML);
	}
	row.setAttribute('data-original', JSON.stringify(originalValues));

	var address = addresses.find(a => a.id == id);
	if (!address) return;

	cells[0].innerHTML = '<input type="text" class="form-control" value="' + escapeHtml(address.street) + '">';
	cells[1].innerHTML = '<input type="text" class="form-control" value="' + escapeHtml(address.building) + '">';

	if (address.cabinet) {
		cells[2].innerHTML = '<input type="text" class="form-control" value="' + escapeHtml(address.cabinet) + '">';
		cells[3].innerHTML = '-';
		cells[4].innerHTML = '-';
		cells[5].innerHTML = '-';
	} else if (address.corridor) {
		cells[2].innerHTML = '-';
		cells[3].innerHTML = '<input type="text" class="form-control" value="' + escapeHtml(address.corridor) + '">';
		cells[4].innerHTML = '<input type="number" class="form-control" min="1" max="5" value="' + (address.floor || '') + '">';
		cells[5].innerHTML = '-';
	} else if (address.service_room) {
		cells[2].innerHTML = '-';
		cells[3].innerHTML = '-';
		cells[4].innerHTML = '-';
		cells[5].innerHTML = '<input type="text" class="form-control" value="' + escapeHtml(address.service_room) + '">';
	} else {
		cells[2].innerHTML = '<input type="text" class="form-control" value="">';
		cells[3].innerHTML = '-';
		cells[4].innerHTML = '-';
		cells[5].innerHTML = '-';
	}

	row.querySelector('.btn-edit').style.display = 'none';
	row.querySelector('.btn-save').style.display = 'inline-block';
	row.querySelector('.btn-cancel').style.display = 'inline-block';
	row.classList.add('edit-mode');
}

function saveAddressRow(id, button) {
	var row = button.closest('tr');
	var inputs = row.querySelectorAll('input');

	var data = {
		street: inputs[0].value.trim(),
		building: inputs[1].value.trim()
	};

	if (inputs.length >= 3 && inputs[2].value.trim()) {
		data.cabinet = inputs[2].value.trim();
	} else if (inputs.length >= 4 && inputs[3].value.trim()) {
		data.corridor = inputs[3].value.trim();
		data.floor = parseInt(inputs[4].value) || null;
	} else if (inputs.length >= 5 && inputs[5].value.trim()) {
		data.service_room = inputs[5].value.trim();
	}

	if (!data.street || !data.building) {
		showAlert(addressAlert, 'Улица и дом обязательны для заполнения', 'danger');
		return;
	}

	fetch('/api/reference/addresses/' + id, {
		method: 'PUT',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(data)
	})
	.then(response => response.json())
	.then(result => {
		fetch('/api/reference/addresses')
			.then(r => r.json())
			.then(data => {
				addresses = data;
				renderAddressesTable();
				populateAddressDropdowns();
				populateWorkstationDropdowns();
				showAlert(null, 'Адрес успешно обновлён!', 'success', true);
			});
	})
	.catch(error => {
		showAlert(addressAlert, error.message || 'Ошибка сети', 'danger');
	});
}

function cancelAddressEdit(id, button) {
	var row = button.closest('tr');
	var originalValues = JSON.parse(row.getAttribute('data-original'));

	var cells = row.querySelectorAll('td');
	for (var i = 0; i < originalValues.length; i++) {
		cells[i].innerHTML = originalValues[i];
	}

	row.querySelector('.btn-edit').style.display = 'inline-block';
	row.querySelector('.btn-save').style.display = 'none';
	row.querySelector('.btn-cancel').style.display = 'none';
	row.classList.remove('edit-mode');
}

function saveEmployee() {
	var fullName = employeeFullName.value.trim();
	var shortName = employeeShortName.value.trim();
	var phoneCity = employeePhoneCity.value.trim();
	var phoneInternal = employeePhoneInternal.value.trim();
	var addressId = parseInt(employeeAddress.value);

	if (!fullName) {
		showAlert(employeeAlert, 'ФИО полностью обязательно для заполнения', 'danger');
		return;
	}
	if (!addressId) {
		showAlert(employeeAlert, 'Адрес обязателен для заполнения', 'danger');
		return;
	}

	var data = {
		full_name: fullName,
		short_name: shortName,
		phone_city: phoneCity,
		phone_internal: phoneInternal,
		address_id: addressId
	};

	fetch('/api/reference/employees', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(data)
	})
	.then(response => response.json())
	.then(result => {
		employeeFullName.value = '';
		employeeShortName.value = '';
		employeePhoneCity.value = '';
		employeePhoneInternal.value = '';
		employeeAddress.value = '';

		fetch('/api/reference/employees')
			.then(r => r.json())
			.then(data => {
				employees = data;
				renderEmployeesTable();
				populateEmployeeDropdowns();
				populateWorkstationDropdowns();
				showAlert(null, 'Сотрудник успешно добавлен!', 'success', true);
			});
	})
	.catch(error => {
		showAlert(employeeAlert, error.message || 'Ошибка сети', 'danger');
	});
}

function deleteEmployee(id) {
	if (!confirm('Вы уверены, что хотите удалить этого сотрудника?')) {
		return;
	}
	fetch('/api/reference/employees/' + id, {
		method: 'DELETE'
	})
	.then(response => response.json())
	.then(result => {
		fetch('/api/reference/employees')
			.then(r => r.json())
			.then(data => {
				employees = data;
				renderEmployeesTable();
				populateEmployeeDropdowns();
				populateWorkstationDropdowns();
				showAlert(null, 'Сотрудник успешно удалён!', 'success', true);
			});
	})
	.catch(error => {
		showAlert(null, error.message || 'Ошибка удаления сотрудника', 'danger', true);
	});
}

function editEmployeeRow(empId, button) {
	var row = button.closest('tr');
	var cells = row.querySelectorAll('td');
	var originalValues = [];
	for (var i = 0; i < cells.length - 1; i++) {
		originalValues.push(cells[i].innerHTML);
	}
	row.setAttribute('data-original', JSON.stringify(originalValues));

	var employee = employees.find(e => e.id == empId);
	if (!employee) return;

	cells[0].innerHTML = '<input type="text" class="form-control" value="' + escapeHtml(employee.full_name) + '">';
	cells[1].innerHTML = '<input type="text" class="form-control" value="' + escapeHtml(employee.short_name || '') + '">';
	cells[2].innerHTML = '<input type="text" class="form-control" value="' + escapeHtml(employee.phone_city || '') + '">';
	cells[3].innerHTML = '<input type="text" class="form-control" value="' + escapeHtml(employee.phone_internal || '') + '">';
	cells[4].innerHTML = '<select class="form-control">' +
		addresses.map(a => '<option value="' + a.id + '"' + (a.id == employee.address_id ? ' selected' : '') + '>' + escapeHtml(a.full_address) + '</option>').join('') +
		'</select>';

	row.querySelector('.btn-edit').style.display = 'none';
	row.querySelector('.btn-save').style.display = 'inline-block';
	row.querySelector('.btn-cancel').style.display = 'inline-block';
	row.classList.add('edit-mode');
}

function saveEmployeeRow(empId, button) {
	var row = button.closest('tr');
	var inputs = row.querySelectorAll('input, select');
	var data = {
		full_name: inputs[0].value.trim(),
		short_name: inputs[1].value.trim(),
		phone_city: inputs[2].value.trim(),
		phone_internal: inputs[3].value.trim(),
		address_id: parseInt(inputs[4].value)
	};

	if (!data.full_name || !data.address_id) {
		showAlert(employeeAlert, 'ФИО и адрес обязательны для заполнения', 'danger');
		return;
	}

	fetch('/api/reference/employees/' + empId, {
		method: 'PUT',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(data)
	})
	.then(response => response.json())
	.then(result => {
		fetch('/api/reference/employees')
			.then(r => r.json())
			.then(data => {
				employees = data;
				renderEmployeesTable();
				populateEmployeeDropdowns();
				populateWorkstationDropdowns();
				showAlert(null, 'Сотрудник успешно обновлён!', 'success', true);
			});
	})
	.catch(error => {
		showAlert(employeeAlert, error.message || 'Ошибка сети', 'danger');
	});
}

function cancelEmployeeEdit(empId, button) {
	var row = button.closest('tr');
	var originalValues = JSON.parse(row.getAttribute('data-original'));
	var cells = row.querySelectorAll('td');
	for (var i = 0; i < originalValues.length; i++) {
		cells[i].innerHTML = originalValues[i];
	}

	row.querySelector('.btn-edit').style.display = 'inline-block';
	row.querySelector('.btn-save').style.display = 'none';
	row.querySelector('.btn-cancel').style.display = 'none';
	row.classList.remove('edit-mode');
}

function saveWorkstation() {
	var addressId = parseInt(document.getElementById('workstation-address-value').value);
	var isVacant = workstationVacant.checked;
	var employeeId = isVacant ? 0 : parseInt(document.getElementById('workstation-employee-value').value);
	var inventoryNumber = workstationInventory.value.trim();
	var serialNumber = workstationSerial.value.trim();
	var seals = workstationSeals.value.trim();
	var monitorCount = parseInt(workstationMonitors.value);
	var replacementDone = workstationReplacement.checked;
	var replacementDate = replacementDone ? workstationReplacementDate.value : '';
	var replacementLetter = replacementDone ? workstationReplacementLetter.value.trim() : '';

	if (!addressId) {
		showAlert(workstationAlert, 'Адрес обязателен для заполнения', 'danger');
		return;
	}
	if (!inventoryNumber) {
		showAlert(workstationAlert, 'Инвентарный номер обязателен для заполнения', 'danger');
		return;
	}
	if (!isVacant && !employeeId) {
		showAlert(workstationAlert, 'Для не вакантного места необходимо выбрать сотрудника', 'danger');
		return;
	}
	if (replacementDone && !replacementDate) {
		showAlert(workstationAlert, 'При отметке о замене необходимо указать дату замены', 'danger');
		return;
	}

	var sealArray = [];
	if (seals) {
		sealArray = seals.split(',').map(s => s.trim()).filter(s => s);
	}

	var data = {
		address_id: addressId,
		is_vacant: isVacant,
		employee_id: employeeId || null,
		inventory_number: inventoryNumber,
		seal_numbers: JSON.stringify(sealArray),
		monitor_count: monitorCount,
		serial_number: serialNumber,
		replacement_done: replacementDone,
		replacement_date: replacementDate || null,
		replacement_letter: replacementLetter || null
	};

	fetch('/api/reference/workstations', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(data)
	})
	.then(response => response.json())
	.then(result => {
		// Очищаем форму
		document.getElementById('workstation-address-input').value = '';
		document.getElementById('workstation-address-value').value = '';
		document.getElementById('workstation-employee-input').value = '';
		document.getElementById('workstation-employee-value').value = '';
		workstationVacant.checked = false;
		toggleWorkstationEmployee();
		workstationInventory.value = '';
		workstationSerial.value = '';
		workstationSeals.value = '';
		workstationMonitors.value = '2';
		workstationReplacement.checked = false;
		toggleWorkstationReplacement();
		workstationReplacementDate.value = '';
		workstationReplacementLetter.value = '';

		// Перезагружаем данные и обновляем таблицу
		// ИСПРАВЛЕНО: обновляем ТОЛЬКО контейнер таблицы, а не всю страницу
		fetch('/api/reference/workstations')
			.then(r => r.json())
			.then(data => {
				workstations = data;
				filteredWorkstations = []; // Сбрасываем фильтрацию
				// Таблица обновится через renderWorkstationsTable() с сохранением прокрутки
				// Пагинация обновится через вызов изнутри renderWorkstationsTable()
				renderWorkstationsTable();
				showAlert(null, 'АРМ успешно добавлен!', 'success', true);
			});
	})
	.catch(error => {
		showAlert(workstationAlert, error.message || 'Ошибка сети', 'danger');
	});
}

function deleteWorkstation(id) {
	if (!confirm('Вы уверены, что хотите удалить это АРМ?')) {
		return;
	}
	fetch('/api/reference/workstations/' + id, {
		method: 'DELETE'
	})
	.then(response => response.json())
	.then(result => {
		fetch('/api/reference/workstations')
			.then(r => r.json())
			.then(data => {
				workstations = data;
				filteredWorkstations = []; // Сбрасываем фильтрацию
				// ИСПРАВЛЕНО: обновляем ТОЛЬКО контейнер таблицы
				// Пагинация обновится через вызов изнутри renderWorkstationsTable()
				renderWorkstationsTable();
				showAlert(null, 'АРМ успешно удалён!', 'success', true);
			});
	})
	.catch(error => {
		showAlert(null, error.message || 'Ошибка удаления АРМ', 'danger', true);
	});
}

function editWorkstationRow(id, button) {
	var row = button.closest('tr');
	var cells = row.querySelectorAll('td');
	var ws = workstations.find(w => w.id == id);
	if (!ws) return;

	cells[0].innerHTML = '<select class="form-control">' +
		addresses.map(a => '<option value="' + a.id + '"' + (a.id == ws.address_id ? ' selected' : '') + '>' + escapeHtml(a.full_address) + '</option>').join('') +
		'</select>';

	cells[1].innerHTML = '<select class="form-control">' +
		'<option value="0">-- Вакантное место --</option>' +
		employees.map(e => '<option value="' + e.id + '"' + (e.id == ws.employee_id ? ' selected' : '') + '>' + escapeHtml(e.short_name || e.full_name) + '</option>').join('') +
		'</select>';

	cells[2].innerHTML = '<input type="text" class="form-control" value="' + escapeHtml(ws.inventory_number) + '">';
	cells[3].innerHTML = '<input type="text" class="form-control" value="' + escapeHtml(ws.serial_number || '') + '">';

	var sealNumbers = '';
	try {
		var seals = JSON.parse(ws.seal_numbers || '[]');
		sealNumbers = seals.join(', ');
	} catch(e) {}
	cells[4].innerHTML = '<input type="text" class="form-control" value="' + escapeHtml(sealNumbers) + '">';

	cells[5].innerHTML = '<select class="form-control">' +
		'<option value="1"' + (ws.monitor_count == 1 ? ' selected' : '') + '>1</option>' +
		'<option value="2"' + (ws.monitor_count == 2 ? ' selected' : '') + '>2</option>' +
		'<option value="3"' + (ws.monitor_count == 3 ? ' selected' : '') + '>3</option>' +
		'<option value="4"' + (ws.monitor_count == 4 ? ' selected' : '') + '>4</option>' +
		'<option value="5"' + (ws.monitor_count == 5 ? ' selected' : '') + '>5</option>' +
		'</select>';

	var isVacant = ws.is_vacant === '1' || ws.is_vacant === 1 || ws.is_vacant === true;
	// ИСПРАВЛЕНО: добавлена подпись "Вакантное" рядом с чекбоксом
	cells[6].innerHTML = '<div style="display:flex;align-items:center;gap:8px;">' +
		'<input type="checkbox" id="ws-edit-vacant-' + id + '" ' + (isVacant ? 'checked' : '') + ' style="margin:0;width:auto;">' +
		'<label for="ws-edit-vacant-' + id + '" style="margin:0;font-size:0.9rem;font-weight:500;color:#555;">Вакантное</label>' +
		'</div>';

	// ИСПРАВЛЕНО: корректная обработка полей замены при редактировании
	var replacementDateValue = ws.replacement_date;
	if (replacementDateValue === null || replacementDateValue === 'null' || replacementDateValue === '<nil>' || replacementDateValue === 'NULL' || replacementDateValue === '') {
		replacementDateValue = '';
	}

	var replacementLetterValue = ws.replacement_letter;
	if (replacementLetterValue === null || replacementLetterValue === 'null' || replacementLetterValue === '<nil>' || replacementLetterValue === 'NULL' || replacementLetterValue === '') {
		replacementLetterValue = '';
	}

	cells[7].innerHTML = '<input type="date" class="form-control" value="' + replacementDateValue + '">';
	cells[8].innerHTML = '<input type="text" class="form-control" value="' + escapeHtml(replacementLetterValue) + '">';

	cells[9].innerHTML = 
		'<button class="btn btn-sm btn-success" onclick="saveWorkstationRow(' + id + ', this)">💾 Сохранить</button> ' +
		'<button class="btn btn-sm btn-secondary" onclick="cancelWorkstationEdit(' + id + ', this)">❌ Отмена</button>';
}

function saveWorkstationRow(id, button) {
	var row = button.closest('tr');
	var inputs = row.querySelectorAll('input, select');
	
	// Индексы элементов формы:
	// [0] - select адреса
	// [1] - select сотрудника
	// [2] - input инвентарного номера
	// [3] - input серийного номера
	// [4] - input пломб
	// [5] - select количества мониторов
	// [6] - checkbox вакантности
	// [7] - input даты замены
	// [8] - input буквы замены
	
	var sealArray = [];
	var sealInput = inputs[4].value.trim();
	if (sealInput) {
		sealArray = sealInput.split(',').map(s => s.trim()).filter(s => s);
	}

	var isVacant = inputs[6].checked;
	var replacementDate = inputs[7].value || null;
	var replacementLetter = inputs[8].value.trim() || null;

	var data = {
		address_id: parseInt(inputs[0].value),
		employee_id: isVacant ? null : (parseInt(inputs[1].value) || null),
		inventory_number: inputs[2].value.trim(),
		serial_number: inputs[3].value.trim(),
		seal_numbers: JSON.stringify(sealArray),
		monitor_count: parseInt(inputs[5].value),
		is_vacant: isVacant,
		replacement_done: replacementDate !== null,
		replacement_date: replacementDate,
		replacement_letter: replacementLetter
	};

	if (!data.inventory_number) {
		showAlert(workstationAlert, 'Инвентарный номер обязателен для заполнения', 'danger');
		return;
	}

	if (!data.is_vacant && !data.employee_id) {
		showAlert(workstationAlert, 'Для не вакантного места необходимо выбрать сотрудника', 'danger');
		return;
	}

	fetch('/api/reference/workstations/' + id, {
		method: 'PUT',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(data)
	})
	.then(response => response.json())
	.then(result => {
		fetch('/api/reference/workstations')
			.then(r => r.json())
			.then(data => {
				workstations = data;
				filteredWorkstations = []; // Сбрасываем фильтрацию
				// ИСПРАВЛЕНО: обновляем ТОЛЬКО контейнер таблицы
				// Пагинация обновится через вызов изнутри renderWorkstationsTable()
				renderWorkstationsTable();
				showAlert(null, 'АРМ успешно обновлён!', 'success', true);
			});
	})
	.catch(error => {
		showAlert(workstationAlert, error.message || 'Ошибка сети', 'danger');
	});
}

function cancelWorkstationEdit(id, button) {
	fetch('/api/reference/workstations')
		.then(r => r.json())
		.then(data => {
			workstations = data;
			filteredWorkstations = []; // Сбрасываем фильтрацию
			// ИСПРАВЛЕНО: обновляем ТОЛЬКО контейнер таблицы
			// Пагинация обновится через вызов изнутри renderWorkstationsTable()
			renderWorkstationsTable();
		});
}

// ==================== ФУНКЦИИ РЕДАКТИРОВАНИЯ ДЛЯ ХОСТОВ ====================
function editHostRow(id, button) {
	var row = button.closest('tr');
	var cells = row.querySelectorAll('td');
	var host = hosts.find(h => h.id == id);
	if (!host) return;

	cells[0].innerHTML = '<select class="form-control">' +
		addresses.map(a => '<option value="' + a.id + '"' + (a.id == host.address_id ? ' selected' : '') + '>' + escapeHtml(a.full_address) + '</option>').join('') +
		'</select>';

	cells[1].innerHTML = '<select class="form-control">' +
		'<option value="0">-- Не назначен --</option>' +
		employees.map(e => '<option value="' + e.id + '"' + (e.id == host.employee_id ? ' selected' : '') + '>' + escapeHtml(e.short_name || e.full_name) + '</option>').join('') +
		'</select>';

	cells[2].innerHTML = '<input type="text" class="form-control" value="' + escapeHtml(host.ip) + '">';
	cells[3].innerHTML = '<input type="number" class="form-control" value="' + host.ssh_port + '" min="1" max="65535">';

	var isEnabled = host.enabled === '1' || host.enabled === 1 || host.enabled === true;
	cells[4].innerHTML = '<div style="display:flex;align-items:center;gap:8px;">' +
		'<input type="checkbox" id="host-edit-enabled-' + id + '" ' + (isEnabled ? 'checked' : '') + ' style="margin:0;width:auto;">' +
		'<label for="host-edit-enabled-' + id + '" style="margin:0;font-size:0.9rem;font-weight:500;color:#555;">Включён</label>' +
		'</div>';

	cells[5].innerHTML = 
		'<button class="btn btn-sm btn-success" onclick="saveHostRow(' + id + ', this)">💾 Сохранить</button> ' +
		'<button class="btn btn-sm btn-secondary" onclick="cancelHostEdit(' + id + ', this)">❌ Отмена</button>';
}

function saveHostRow(id, button) {
	var row = button.closest('tr');
	var inputs = row.querySelectorAll('input, select');

	var data = {
		address_id: parseInt(inputs[0].value),
		employee_id: parseInt(inputs[1].value) || null,
		ip: inputs[2].value.trim(),
		ssh_port: parseInt(inputs[3].value),
		enabled: inputs[4].querySelector('input[type="checkbox"]').checked
	};

	if (!data.address_id) {
		showAlert(hostAlert, 'Адрес обязателен для заполнения', 'danger');
		return;
	}
	if (!data.ip) {
		showAlert(hostAlert, 'IP адрес обязателен для заполнения', 'danger');
		return;
	}

	fetch('/api/reference/hosts/' + id, {
		method: 'PUT',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(data)
	})
	.then(response => response.json())
	.then(result => {
		fetch('/api/reference/hosts')
			.then(r => r.json())
			.then(data => {
				hosts = data;
				renderHostsTable();
				showAlert(null, 'Хост успешно обновлён!', 'success', true);
			});
	})
	.catch(error => {
		showAlert(hostAlert, error.message || 'Ошибка сети', 'danger');
	});
}

function cancelHostEdit(id, button) {
	fetch('/api/reference/hosts')
		.then(r => r.json())
		.then(data => {
			hosts = data;
			renderHostsTable();
		});
}

// ==================== ФУНКЦИИ РЕДАКТИРОВАНИЯ ДЛЯ СЕТЕВОГО ОБОРУДОВАНИЯ ====================
function editNetworkEquipmentRow(id, button) {
	var row = button.closest('tr');
	var cells = row.querySelectorAll('td');
	var eq = networkEquipment.find(e => e.id == id);
	if (!eq) return;

	cells[0].innerHTML = '<select class="form-control">' +
		addresses.map(a => '<option value="' + a.id + '"' + (a.id == eq.address_id ? ' selected' : '') + '>' + escapeHtml(a.full_address) + '</option>').join('') +
		'</select>';

	cells[1].innerHTML = '<input type="text" class="form-control" value="' + escapeHtml(eq.category) + '">';
	cells[2].innerHTML = '<input type="text" class="form-control" value="' + escapeHtml(eq.type) + '">';
	cells[3].innerHTML = '<input type="text" class="form-control" value="' + escapeHtml(eq.model) + '">';
	cells[4].innerHTML = '<input type="number" class="form-control" value="' + eq.port_count + '" min="1">';

	cells[5].innerHTML = 
		'<button class="btn btn-sm btn-success" onclick="saveNetworkEquipmentRow(' + id + ', this)">💾 Сохранить</button> ' +
		'<button class="btn btn-sm btn-secondary" onclick="cancelNetworkEquipmentEdit(' + id + ', this)">❌ Отмена</button>';
}

function saveNetworkEquipmentRow(id, button) {
	var row = button.closest('tr');
	var inputs = row.querySelectorAll('input, select');

	var data = {
		address_id: parseInt(inputs[0].value),
		category: inputs[1].value.trim(),
		type: inputs[2].value.trim(),
		model: inputs[3].value.trim(),
		port_count: parseInt(inputs[4].value)
	};

	if (!data.address_id) {
		showAlert(networkEquipmentAlert, 'Адрес обязателен для заполнения', 'danger');
		return;
	}
	if (!data.category) {
		showAlert(networkEquipmentAlert, 'Категория обязательна для заполнения', 'danger');
		return;
	}
	if (!data.type) {
		showAlert(networkEquipmentAlert, 'Тип обязателен для заполнения', 'danger');
		return;
	}
	if (!data.model) {
		showAlert(networkEquipmentAlert, 'Модель обязательна для заполнения', 'danger');
		return;
	}
	if (!data.port_count || data.port_count <= 0) {
		showAlert(networkEquipmentAlert, 'Число портов должно быть больше 0', 'danger');
		return;
	}

	fetch('/api/reference/network-equipment/' + id, {
		method: 'PUT',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(data)
	})
	.then(response => response.json())
	.then(result => {
		fetch('/api/reference/network-equipment')
			.then(r => r.json())
			.then(data => {
				networkEquipment = data;
				renderNetworkEquipmentTable();
				showAlert(null, 'Оборудование успешно обновлено!', 'success', true);
			});
	})
	.catch(error => {
		showAlert(networkEquipmentAlert, error.message || 'Ошибка сети', 'danger');
	});
}

function cancelNetworkEquipmentEdit(id, button) {
	fetch('/api/reference/network-equipment')
		.then(r => r.json())
		.then(data => {
			networkEquipment = data;
			renderNetworkEquipmentTable();
		});
}

// ==================== ФУНКЦИИ РЕДАКТИРОВАНИЯ ДЛЯ МФУ ====================
function editNetworkMFPRow(id, button) {
	var row = button.closest('tr');
	var cells = row.querySelectorAll('td');
	var mfp = networkMFPs.find(m => m.id == id);
	if (!mfp) return;

	cells[0].innerHTML = '<select class="form-control">' +
		addresses.map(a => '<option value="' + a.id + '"' + (a.id == mfp.address_id ? ' selected' : '') + '>' + escapeHtml(a.full_address) + '</option>').join('') +
		'</select>';

	cells[1].innerHTML = '<input type="text" class="form-control" value="' + escapeHtml(mfp.model) + '">';
	cells[2].innerHTML = '<input type="text" class="form-control" value="' + escapeHtml(mfp.ip) + '">';
	cells[3].innerHTML = '<input type="text" class="form-control" value="' + escapeHtml(mfp.hostname || '') + '">';
	cells[4].innerHTML = '<input type="text" class="form-control" value="' + escapeHtml(mfp.serial_number) + '">';

	cells[5].innerHTML = 
		'<button class="btn btn-sm btn-success" onclick="saveNetworkMFPRow(' + id + ', this)">💾 Сохранить</button> ' +
		'<button class="btn btn-sm btn-secondary" onclick="cancelNetworkMFPEdit(' + id + ', this)">❌ Отмена</button>';
}

function saveNetworkMFPRow(id, button) {
	var row = button.closest('tr');
	var inputs = row.querySelectorAll('input, select');

	var data = {
		address_id: parseInt(inputs[0].value),
		model: inputs[1].value.trim(),
		ip: inputs[2].value.trim(),
		hostname: inputs[3].value.trim() || null,
		serial_number: inputs[4].value.trim()
	};

	if (!data.address_id) {
		showAlert(networkMFPAlert, 'Адрес обязателен для заполнения', 'danger');
		return;
	}
	if (!data.model) {
		showAlert(networkMFPAlert, 'Модель обязательна для заполнения', 'danger');
		return;
	}
	if (!data.ip) {
		showAlert(networkMFPAlert, 'IP адрес обязателен для заполнения', 'danger');
		return;
	}
	if (!data.serial_number) {
		showAlert(networkMFPAlert, 'Серийный номер обязателен для заполнения', 'danger');
		return;
	}

	fetch('/api/reference/network-mfps/' + id, {
		method: 'PUT',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(data)
	})
	.then(response => response.json())
	.then(result => {
		fetch('/api/reference/network-mfps')
			.then(r => r.json())
			.then(data => {
				networkMFPs = data;
				renderNetworkMFPsTable();
				showAlert(null, 'МФУ успешно обновлено!', 'success', true);
			});
	})
	.catch(error => {
		showAlert(networkMFPAlert, error.message || 'Ошибка сети', 'danger');
	});
}

function cancelNetworkMFPEdit(id, button) {
	fetch('/api/reference/network-mfps')
		.then(r => r.json())
		.then(data => {
			networkMFPs = data;
			renderNetworkMFPsTable();
		});
}

// ==================== ФУНКЦИИ РЕДАКТИРОВАНИЯ ДЛЯ IP-ТЕЛЕФОНОВ ====================
function editIPPhoneRow(id, button) {
	var row = button.closest('tr');
	var cells = row.querySelectorAll('td');
	var phone = ipPhones.find(p => p.id == id);
	if (!phone) return;

	cells[0].innerHTML = '<select class="form-control">' +
		addresses.map(a => '<option value="' + a.id + '"' + (a.id == phone.address_id ? ' selected' : '') + '>' + escapeHtml(a.full_address) + '</option>').join('') +
		'</select>';

	cells[1].innerHTML = '<input type="text" class="form-control" value="' + escapeHtml(phone.employee_full_name) + '">';
	cells[2].innerHTML = '<input type="text" class="form-control" value="' + escapeHtml(phone.ip) + '">';

	cells[3].innerHTML = 
		'<button class="btn btn-sm btn-success" onclick="saveIPPhoneRow(' + id + ', this)">💾 Сохранить</button> ' +
		'<button class="btn btn-sm btn-secondary" onclick="cancelIPPhoneEdit(' + id + ', this)">❌ Отмена</button>';
}

function saveIPPhoneRow(id, button) {
	var row = button.closest('tr');
	var inputs = row.querySelectorAll('input, select');

	var data = {
		address_id: parseInt(inputs[0].value),
		employee_full_name: inputs[1].value.trim(),
		ip: inputs[2].value.trim()
	};

	if (!data.address_id) {
		showAlert(ipPhoneAlert, 'Адрес обязателен для заполнения', 'danger');
		return;
	}
	if (!data.employee_full_name) {
		showAlert(ipPhoneAlert, 'ФИО сотрудника обязательно для заполнения', 'danger');
		return;
	}
	if (!data.ip) {
		showAlert(ipPhoneAlert, 'IP адрес обязателен для заполнения', 'danger');
		return;
	}

	fetch('/api/reference/ip-phones/' + id, {
		method: 'PUT',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(data)
	})
	.then(response => response.json())
	.then(result => {
		fetch('/api/reference/ip-phones')
			.then(r => r.json())
			.then(data => {
				ipPhones = data;
				renderIPPhonesTable();
				showAlert(null, 'IP телефон успешно обновлён!', 'success', true);
			});
	})
	.catch(error => {
		showAlert(ipPhoneAlert, error.message || 'Ошибка сети', 'danger');
	});
}

function cancelIPPhoneEdit(id, button) {
	fetch('/api/reference/ip-phones')
		.then(r => r.json())
		.then(data => {
			ipPhones = data;
			renderIPPhonesTable();
		});
}

function saveHost() {
	var addressId = parseInt(hostAddress.value);
	var employeeId = parseInt(hostEmployee.value) || null;
	var ip = hostIP.value.trim();
	var sshPort = parseInt(hostSSHPort.value) || 22;
	var enabled = hostEnabled.checked;

	if (!addressId) {
		showAlert(hostAlert, 'Адрес обязателен для заполнения', 'danger');
		return;
	}
	if (!ip) {
		showAlert(hostAlert, 'IP адрес обязателен для заполнения', 'danger');
		return;
	}

	var data = {
		address_id: addressId,
		employee_id: employeeId,
		ip: ip,
		ssh_port: sshPort,
		enabled: enabled
	};

	fetch('/api/reference/hosts', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(data)
	})
	.then(response => response.json())
	.then(result => {
		hostAddress.value = '';
		hostEmployee.value = '';
		hostIP.value = '';
		hostSSHPort.value = '22';
		hostEnabled.checked = true;

		fetch('/api/reference/hosts')
			.then(r => r.json())
			.then(data => {
				hosts = data;
				renderHostsTable();
				showAlert(null, 'Хост успешно добавлен!', 'success', true);
			});
	})
	.catch(error => {
		showAlert(hostAlert, error.message || 'Ошибка сети', 'danger');
	});
}

function deleteHost(id) {
	if (!confirm('Вы уверены, что хотите удалить этот хост?')) {
		return;
	}
	fetch('/api/reference/hosts/' + id, {
		method: 'DELETE'
	})
	.then(response => response.json())
	.then(result => {
		fetch('/api/reference/hosts')
			.then(r => r.json())
			.then(data => {
				hosts = data;
				renderHostsTable();
				showAlert(null, 'Хост успешно удалён!', 'success', true);
			});
	})
	.catch(error => {
		showAlert(null, error.message || 'Ошибка удаления хоста', 'danger', true);
	});
}

function saveNetworkEquipment() {
	var addressId = parseInt(networkEquipmentAddress.value);
	var category = networkEquipmentCategory.value.trim();
	var type = networkEquipmentType.value.trim();
	var model = networkEquipmentModel.value.trim();
	var portCount = parseInt(networkEquipmentPorts.value);

	if (!addressId) {
		showAlert(networkEquipmentAlert, 'Адрес обязателен для заполнения', 'danger');
		return;
	}
	if (!category) {
		showAlert(networkEquipmentAlert, 'Категория обязательна для заполнения', 'danger');
		return;
	}
	if (!type) {
		showAlert(networkEquipmentAlert, 'Тип обязателен для заполнения', 'danger');
		return;
	}
	if (!model) {
		showAlert(networkEquipmentAlert, 'Модель обязательна для заполнения', 'danger');
		return;
	}
	if (!portCount || portCount <= 0) {
		showAlert(networkEquipmentAlert, 'Число портов должно быть больше 0', 'danger');
		return;
	}

	var data = {
		address_id: addressId,
		category: category,
		type: type,
		model: model,
		port_count: portCount
	};

	fetch('/api/reference/network-equipment', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(data)
	})
	.then(response => response.json())
	.then(result => {
		networkEquipmentAddress.value = '';
		networkEquipmentCategory.value = '';
		networkEquipmentType.value = '';
		networkEquipmentModel.value = '';
		networkEquipmentPorts.value = '24';

		fetch('/api/reference/network-equipment')
			.then(r => r.json())
			.then(data => {
				networkEquipment = data;
				renderNetworkEquipmentTable();
				showAlert(null, 'Сетевое оборудование успешно добавлено!', 'success', true);
			});
	})
	.catch(error => {
		showAlert(networkEquipmentAlert, error.message || 'Ошибка сети', 'danger');
	});
}

function deleteNetworkEquipment(id) {
	if (!confirm('Вы уверены, что хотите удалить это оборудование?')) {
		return;
	}
	fetch('/api/reference/network-equipment/' + id, {
		method: 'DELETE'
	})
	.then(response => response.json())
	.then(result => {
		fetch('/api/reference/network-equipment')
			.then(r => r.json())
			.then(data => {
				networkEquipment = data;
				renderNetworkEquipmentTable();
				showAlert(null, 'Сетевое оборудование успешно удалено!', 'success', true);
			});
	})
	.catch(error => {
		showAlert(null, error.message || 'Ошибка удаления оборудования', 'danger', true);
	});
}

function saveNetworkMFP() {
	var addressId = parseInt(networkMFPAddress.value);
	var model = networkMFPModel.value.trim();
	var ip = networkMFPIP.value.trim();
	var hostname = networkMFPHostname.value.trim();
	var serialNumber = networkMFPSerial.value.trim();

	if (!addressId) {
		showAlert(networkMFPAlert, 'Адрес обязателен для заполнения', 'danger');
		return;
	}
	if (!model) {
		showAlert(networkMFPAlert, 'Модель обязательна для заполнения', 'danger');
		return;
	}
	if (!ip) {
		showAlert(networkMFPAlert, 'IP адрес обязателен для заполнения', 'danger');
		return;
	}
	if (!serialNumber) {
		showAlert(networkMFPAlert, 'Серийный номер обязателен для заполнения', 'danger');
		return;
	}

	var data = {
		address_id: addressId,
		model: model,
		ip: ip,
		hostname: hostname || null,
		serial_number: serialNumber
	};

	fetch('/api/reference/network-mfps', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(data)
	})
	.then(response => response.json())
	.then(result => {
		networkMFPAddress.value = '';
		networkMFPModel.value = '';
		networkMFPIP.value = '';
		networkMFPHostname.value = '';
		networkMFPSerial.value = '';

		fetch('/api/reference/network-mfps')
			.then(r => r.json())
			.then(data => {
				networkMFPs = data;
				renderNetworkMFPsTable();
				showAlert(null, 'Сетевое МФУ успешно добавлено!', 'success', true);
			});
	})
	.catch(error => {
		showAlert(networkMFPAlert, error.message || 'Ошибка сети', 'danger');
	});
}

function deleteNetworkMFP(id) {
	if (!confirm('Вы уверены, что хотите удалить это МФУ?')) {
		return;
	}
	fetch('/api/reference/network-mfps/' + id, {
		method: 'DELETE'
	})
	.then(response => response.json())
	.then(result => {
		fetch('/api/reference/network-mfps')
			.then(r => r.json())
			.then(data => {
				networkMFPs = data;
				renderNetworkMFPsTable();
				showAlert(null, 'Сетевое МФУ успешно удалено!', 'success', true);
			});
	})
	.catch(error => {
		showAlert(null, error.message || 'Ошибка удаления МФУ', 'danger', true);
	});
}

function saveIPPhone() {
	var addressId = parseInt(ipPhoneAddress.value);
	var employeeFullName = ipPhoneEmployee.value.trim();
	var ip = ipPhoneIP.value.trim();

	if (!addressId) {
		showAlert(ipPhoneAlert, 'Адрес обязателен для заполнения', 'danger');
		return;
	}
	if (!employeeFullName) {
		showAlert(ipPhoneAlert, 'ФИО сотрудника обязательно для заполнения', 'danger');
		return;
	}
	if (!ip) {
		showAlert(ipPhoneAlert, 'IP адрес обязателен для заполнения', 'danger');
		return;
	}

	var data = {
		address_id: addressId,
		employee_full_name: employeeFullName,
		ip: ip
	};

	fetch('/api/reference/ip-phones', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(data)
	})
	.then(response => response.json())
	.then(result => {
		ipPhoneAddress.value = '';
		ipPhoneEmployee.value = '';
		ipPhoneIP.value = '';

		fetch('/api/reference/ip-phones')
			.then(r => r.json())
			.then(data => {
				ipPhones = data;
				renderIPPhonesTable();
				showAlert(null, 'IP телефон успешно добавлен!', 'success', true);
			});
	})
	.catch(error => {
		showAlert(ipPhoneAlert, error.message || 'Ошибка сети', 'danger');
	});
}

function deleteIPPhone(id) {
	if (!confirm('Вы уверены, что хотите удалить этот телефон?')) {
		return;
	}
	fetch('/api/reference/ip-phones/' + id, {
		method: 'DELETE'
	})
	.then(response => response.json())
	.then(result => {
		fetch('/api/reference/ip-phones')
			.then(r => r.json())
			.then(data => {
				ipPhones = data;
				renderIPPhonesTable();
				showAlert(null, 'IP телефон успешно удалён!', 'success', true);
			});
	})
	.catch(error => {
		showAlert(null, error.message || 'Ошибка удаления телефона', 'danger', true);
	});
}

// ==================== ИНИЦИАЛИЗАЦИЯ ОБРАБОТЧИКОВ СОБЫТИЙ ДЛЯ СПРАВОЧНИКОВ ====================
document.addEventListener('DOMContentLoaded', function() {
	if (saveAddressBtn) saveAddressBtn.addEventListener('click', saveAddress);
	if (saveEmployeeBtn) saveEmployeeBtn.addEventListener('click', saveEmployee);
	if (saveWorkstationBtn) saveWorkstationBtn.addEventListener('click', saveWorkstation);
	if (saveHostBtn) saveHostBtn.addEventListener('click', saveHost);
	if (saveNetworkEquipmentBtn) saveNetworkEquipmentBtn.addEventListener('click', saveNetworkEquipment);
	if (saveNetworkMFPBtn) saveNetworkMFPBtn.addEventListener('click', saveNetworkMFP);
	if (saveIPPhoneBtn) saveIPPhoneBtn.addEventListener('click', saveIPPhone);
	
	// Инициализация поиска в форме АРМ
	initWorkstationFormSearch();
	
	// Обработчик изменения размера страницы (ИСПРАВЛЕНО: вызов без условия)
	var pageSizeSelect = document.getElementById('workstations-page-size');
	if (pageSizeSelect) {
		pageSizeSelect.addEventListener('change', function() {
			// ИСПРАВЛЕНО: всегда вызываем обновление
			changeWorkstationsPage(null, 1);
		});
	}
});

// Инициализация поиска в форме АРМ
function initWorkstationFormSearch() {
	// Поиск для адреса
	var addrInput = document.getElementById('workstation-address-input');
	var addrOpts = document.getElementById('workstation-address-options');
	var addrVal = document.getElementById('workstation-address-value');
	
	if (addrInput) {
		addrInput.addEventListener('focus', function() {
			filterOptions(addresses, addrOpts, addrInput.value, function(addr) {
				return addr.full_address;
			});
			addrOpts.style.display = 'block';
		});
		
		addrInput.addEventListener('input', function() {
			filterOptions(addresses, addrOpts, this.value, function(addr) {
				return addr.full_address;
			});
		});
		
		addrInput.addEventListener('blur', function() {
			setTimeout(function() { addrOpts.style.display = 'none'; }, 200);
		});
	}
	
	// Поиск для сотрудника
	var empInput = document.getElementById('workstation-employee-input');
	var empOpts = document.getElementById('workstation-employee-options');
	var empVal = document.getElementById('workstation-employee-value');
	
	if (empInput) {
		empInput.addEventListener('focus', function() {
			filterOptions(employees, empOpts, empInput.value, function(emp) {
				return emp.short_name || emp.full_name;
			});
			empOpts.style.display = 'block';
		});
		
		empInput.addEventListener('input', function() {
			filterOptions(employees, empOpts, this.value, function(emp) {
				return emp.short_name || emp.full_name;
			});
		});
		
		empInput.addEventListener('blur', function() {
			setTimeout(function() { empOpts.style.display = 'none'; }, 200);
		});
	}
}

function filterOptions(items, container, query, getText) {
	container.innerHTML = '';
	var filtered = items.filter(function(item) {
		return getText(item).toLowerCase().includes(query.toLowerCase());
	});
	
	filtered.forEach(function(item) {
		var div = document.createElement('div');
		div.className = 'searchable-option';
		div.style.padding = '8px 12px';
		div.style.cursor = 'pointer';
		div.style.borderBottom = '1px solid #f0f0f0';
		div.textContent = getText(item);
		div.addEventListener('click', function() {
			var inputId = container.id.replace('-options', '-input');
			var valueId = container.id.replace('-options', '-value');
			document.getElementById(inputId).value = getText(item);
			document.getElementById(valueId).value = item.id;
			container.style.display = 'none';
		});
		div.addEventListener('mouseenter', function() { this.style.backgroundColor = '#f5f5f5'; });
		div.addEventListener('mouseleave', function() { this.style.backgroundColor = ''; });
		container.appendChild(div);
	});
	
	if (filtered.length === 0) {
		var empty = document.createElement('div');
		empty.className = 'searchable-option';
		empty.style.padding = '8px 12px';
		empty.style.color = '#999';
		empty.textContent = 'Ничего не найдено';
		container.appendChild(empty);
	}
}
`