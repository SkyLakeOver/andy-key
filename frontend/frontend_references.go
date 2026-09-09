package frontend

// referencesJS содержит все функции для справочников (адреса, сотрудники, АРМ, хосты, оборудование, МФУ, телефоны)
var referencesJS = `
// ==================== УНИВЕРСАЛЬНЫЙ КЛАСС ДЛЯ ТАБЛИЦ СПРАВОЧНИКОВ ====================
class ReferenceTable {
    constructor(config) {
        this.name = config.name;
        this.tableElement = document.getElementById(config.tableId);
        this.data = [];
        this.filteredData = [];
        this.sortColumn = '';
        this.sortDirection = 'asc';
        this.currentPage = 1;
        this.pageSize = 25;
        this.columns = config.columns || [];
        this.searchInputs = [];
        this.init();
    }

    init() {
            console.error('Table element not found for:', this.name);
            return;
        }
        this.setupSearchInputs();
        this.setupEventListeners();
    }

    setupSearchInputs() {
        const inputs = this.tableElement.querySelectorAll('.search-row input');
        this.searchInputs = Array.from(inputs);
        this.searchInputs.forEach((input, index) => {
            input.addEventListener('input', () => this.filterData());
        });
    }

    setData(newData) {
        this.data = Array.isArray(newData) ? newData : [];
        this.filteredData = [...this.data];
        this.currentPage = 1;
        this.sortColumn = '';
        this.sortDirection = 'asc';
        this.clearSearchInputs();
        this.render();
    }

    clearSearchInputs() {
        this.searchInputs.forEach(input => input.value = '');
    }

    resetStateOnly() {
        this.sortColumn = '';
        this.sortDirection = 'asc';
        this.currentPage = 1;
        this.clearSearchInputs();
        this.filteredData = [...this.data];
        this.render();
    }

    filterData() {
        this.filteredData = this.data.filter(item => {
            for (let i = 0; i < this.searchInputs.length; i++) {
                const query = this.searchInputs[i].value.toLowerCase().trim();
                if (query) {
                    const columnName = this.columns[i];
                    const value = (item[columnName] || '').toString().toLowerCase();
                }
            }
            return true;
        });
        this.currentPage = 1;
        this.render();
    }

    sort(column) {
        if (this.sortColumn === column) {
            this.sortDirection = this.sortDirection === 'asc' ? 'desc' : 'asc';
        } else {
            this.sortColumn = column;
            this.sortDirection = 'asc';
        }
        this.render();
    }

    render() {
        const container = document.querySelector('.scrollable-table-container');
        const scrollTop = container ? container.scrollTop : 0;

        if (this.filteredData.length === 0) {
            this.tableElement.innerHTML = '<tr><td colspan="' + (this.columns.length + 1) + '" class="empty-state">Записи не найдены</td></tr>';
            this.updateInfo();
            this.updatePagination(1, 1);
            return;
        }

        let sorted = [...this.filteredData];
        if (this.sortColumn) {
            sorted.sort((a, b) => {
                const valA = (a[this.sortColumn] || '').toString().toLowerCase();
                const valB = (b[this.sortColumn] || '').toString().toLowerCase();
                if (valA < valB) return this.sortDirection === 'asc' ? -1 : 1;
                if (valA > valB) return this.sortDirection === 'asc' ? 1 : -1;
                return 0;
            });
        }

        const totalPages = this.pageSize === 0 ? 1 : Math.ceil(sorted.length / this.pageSize);
        if (this.currentPage > totalPages) this.currentPage = totalPages;
        this.updatePagination(this.currentPage, totalPages);

        const start = this.pageSize === 0 ? 0 : (this.currentPage - 1) * this.pageSize;
        const end = this.pageSize === 0 ? sorted.length : start + this.pageSize;
        const pageItems = this.pageSize === 0 ? sorted : sorted.slice(start, end);

        this.renderRows(pageItems);
        this.updateInfo(start, end, sorted.length);
        if (container) container.scrollTop = scrollTop;
    }

    renderRows(items) {
        let html = '';
        items.forEach(item => {
            html += '<tr data-id="' + this.escapeHtml(String(item.id)) + '">';
            this.columns.forEach(col => {
                html += '<td>' + this.escapeHtml(String(item[col] || '-')) + '</td>';
            });
            html += '<td>' + this.renderActions(item) + '</td></tr>';
        });
        this.tableElement.innerHTML = html;
    }

    renderActions(item) {
        return '<button class="btn btn-sm btn-edit edit-btn" onclick="editRow(\'' + this.name + '\', ' + item.id + ', this)">✏️</button>' +
               '<button class="btn btn-sm btn-save edit-btn" onclick="saveRow(\'' + this.name + '\', ' + item.id + ', this)" style="display:none;">💾</button>' +
               '<button class="btn btn-sm btn-cancel edit-btn" onclick="cancelEdit(\'' + this.name + '\', ' + item.id + ', this)" style="display:none;">❌</button>' +
               '<button class="btn btn-sm btn-danger" onclick="deleteRow(\'' + this.name + '\', ' + item.id + ')">Удалить</button>';
    }

    updatePagination(currentPage, totalPages) {
        const pagination = document.getElementById(this.name + '-pagination');
        let html = '<li class="page-item ' + (currentPage === 1 ? 'disabled' : '') + '"><a class="page-link" href="#" onclick="tables[\'' + this.name + '\'].changePage(event, 'prev')">&laquo;</a></li>';
        for (let i = 1; i <= totalPages && i <= 5; i++) {
            html += '<li class="page-item ' + (i === currentPage ? 'active' : '') + '"><a class="page-link" href="#" onclick="tables[\'' + this.name + '\'].changePage(event, ' + i + ')">' + i + '</a></li>';
        }
        html += '<li class="page-item ' + (currentPage === totalPages ? 'disabled' : '') + '"><a class="page-link" href="#" onclick="tables[\'' + this.name + '\'].changePage(event, 'next')">&raquo;</a></li>';
        pagination.innerHTML = html;
    }

    changePage(event, direction) {
        if (event && event.preventDefault) event.preventDefault();
        const totalPages = this.pageSize === 0 ? 1 : Math.ceil(this.filteredData.length / this.pageSize);
        let newPage = this.currentPage;
        if (direction === 'prev') newPage = Math.max(1, this.currentPage - 1);
        else if (direction === 'next') newPage = Math.min(totalPages, this.currentPage + 1);
        else if (typeof direction === 'number') newPage = Math.max(1, Math.min(totalPages, direction));
        if (newPage !== this.currentPage) {
            this.currentPage = newPage;
            this.render();
        }
    }

    updateInfo(start, end, total) {
        const infoEl = document.getElementById(this.name + '-info');
        if (infoEl) {
            if (this.filteredData.length === 0) infoEl.textContent = 'Показано 0 из ' + this.data.length + ' записей';
            else {
                const displayEnd = this.pageSize === 0 ? this.filteredData.length : Math.min(end, this.filteredData.length);
                infoEl.textContent = 'Показано ' + (start + 1) + '-' + displayEnd + ' из ' + this.filteredData.length + ' записей (всего: ' + this.data.length + ')';
            }
        }
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
}

var tables = {};

document.addEventListener('DOMContentLoaded', function() {
    tables['addresses'] = new ReferenceTable({name: 'addresses', tableId: 'addresses-table-body', columns: ['street', 'building', 'cabinet', 'corridor', 'floor', 'service_room']});
    tables['employees'] = new ReferenceTable({name: 'employees', tableId: 'employees-table-body', columns: ['full_name', 'short_name', 'phone_city', 'phone_internal', 'full_address']});
    tables['workstations'] = new ReferenceTable({name: 'workstations', tableId: 'workstations-table-body', columns: ['name', 'ip_address', 'mac_address', 'os', 'user_full_name', 'address_full']});
    tables['hosts'] = new ReferenceTable({name: 'hosts', tableId: 'hosts-table-body', columns: ['name', 'ip_address', 'mac_address', 'description']});
    tables['network_equipment'] = new ReferenceTable({name: 'network_equipment', tableId: 'network-equipment-table-body', columns: ['name', 'ip_address', 'mac_address', 'model', 'location']});
    tables['network_mfps'] = new ReferenceTable({name: 'network_mfps', tableId: 'network-mfps-table-body', columns: ['name', 'ip_address', 'mac_address', 'model', 'location']});
    tables['ip_phones'] = new ReferenceTable({name: 'ip_phones', tableId: 'ip-phones-table-body', columns: ['name', 'ip_address', 'mac_address', 'model', 'phone_number']});

    document.querySelectorAll('a[data-toggle="tab"]').forEach(tab => {
        tab.addEventListener('shown.bs.tab', function(event) {
            const previousTab = event.relatedTarget;
            if (previousTab) {
                const previousTabId = previousTab.getAttribute('href');
                if (previousTabId) {
                    const refName = previousTabId.replace('#references-', '');
                    if (tables[refName]) tables[refName].resetStateOnly();
                }
            }
        });
    });

    loadReferencesData();
});

function loadReferencesData() {
    fetch('/api/references/all')
        .then(response => response.json())
        .then(data => {
            if (data.addresses) tables['addresses'].setData(data.addresses);
            if (data.employees) tables['employees'].setData(data.employees);
            if (data.workstations) tables['workstations'].setData(data.workstations);
            if (data.hosts) tables['hosts'].setData(data.hosts);
            if (data.network_equipment) tables['network_equipment'].setData(data.network_equipment);
            if (data.network_mfps) tables['network_mfps'].setData(data.network_mfps);
            if (data.ip_phones) tables['ip_phones'].setData(data.ip_phones);
        })
        .catch(error => console.error('Error loading references data:', error));
}

// Заглушки для функций редактирования
function editAddressRow(id, btn) { console.log('editAddressRow', id); }
function saveAddressRow(id, btn) { console.log('saveAddressRow', id); }
function cancelAddressEdit(id, btn) { console.log('cancelAddressEdit', id); }
function deleteAddress(id) { console.log('deleteAddress', id); }
function editEmployeeRow(id, btn) { console.log('editEmployeeRow', id); }
function saveEmployeeRow(id, btn) { console.log('saveEmployeeRow', id); }
function cancelEmployeeEdit(id, btn) { console.log('cancelEmployeeEdit', id); }
function deleteEmployee(id) { console.log('deleteEmployee', id); }
function editWorkstationRow(id, btn) { console.log('editWorkstationRow', id); }
function saveWorkstationRow(id, btn) { console.log('saveWorkstationRow', id); }
function cancelWorkstationEdit(id, btn) { console.log('cancelWorkstationEdit', id); }
function deleteWorkstation(id) { console.log('deleteWorkstation', id); }
function editHostRow(id, btn) { console.log('editHostRow', id); }
function saveHostRow(id, btn) { console.log('saveHostRow', id); }
function cancelHostEdit(id, btn) { console.log('cancelHostEdit', id); }
function deleteHost(id) { console.log('deleteHost', id); }
function editNetworkEquipmentRow(id, btn) { console.log('editNetworkEquipmentRow', id); }
function saveNetworkEquipmentRow(id, btn) { console.log('saveNetworkEquipmentRow', id); }
function cancelNetworkEquipmentEdit(id, btn) { console.log('cancelNetworkEquipmentEdit', id); }
function deleteNetworkEquipment(id) { console.log('deleteNetworkEquipment', id); }
function editNetworkMFPRow(id, btn) { console.log('editNetworkMFPRow', id); }
function saveNetworkMFPRow(id, btn) { console.log('saveNetworkMFPRow', id); }
function cancelNetworkMFPEdit(id, btn) { console.log('cancelNetworkMFPEdit', id); }
function deleteNetworkMFP(id) { console.log('deleteNetworkMFP', id); }
function editIPPhoneRow(id, btn) { console.log('editIPPhoneRow', id); }
function saveIPPhoneRow(id, btn) { console.log('saveIPPhoneRow', id); }
function cancelIPPhoneEdit(id, btn) { console.log('cancelIPPhoneEdit', id); }
function deleteIPPhone(id) { console.log('deleteIPPhone', id); }
`
